package apis

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/qlc"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
)

type QGasSwapAPI struct {
	eth       *eth.Transaction
	qlc       *qlc.Transaction
	store     *gorm.DB
	cfg       *config.Config
	ctx       context.Context
	ownerAddr qlctypes.Address
	addrPool  *AddressPool
	signer    *signer.SignerClient
	logger    *zap.SugaredLogger
}

func NewQGasSwapAPI(ctx context.Context, cfg *config.Config, q *qlc.Transaction, e *eth.Transaction, signer *signer.SignerClient, s *gorm.DB) *QGasSwapAPI {
	api := &QGasSwapAPI{
		cfg:    cfg,
		eth:    e,
		qlc:    q,
		ctx:    ctx,
		store:  s,
		signer: signer,
		logger: log.NewLogger("api/qgas_pledge"),
	}
	address, err := qlctypes.HexToAddress(cfg.QlcCfg.OwnerAddress)
	if err != nil {
		api.logger.Fatal(err)
	}
	api.ownerAddr = address
	api.addrPool = AddressPools(address)
	return api
}

type QGasPledgeParam struct {
	PledgeAddress     qlctypes.Address
	Amount            qlctypes.Balance
	Erc20ReceiverAddr string
}

func (g *QGasSwapAPI) GetPledgeBlock(ctx context.Context, params *pb.QGasPledgeRequest) (*pb.StateBlock, error) {
	g.logger.Infof("QGas Pledge ......... (%s) ", params)
	if params.GetPledgeAddress() == "" || params.Erc20ReceiverAddr == "" || params.GetAmount() <= 0 {
		return nil, errors.New("error params")
	}
	pledgeAddress, err := qlctypes.HexToAddress(params.GetPledgeAddress())
	if err != nil {
		return nil, fmt.Errorf("invalid address, %s", params.GetPledgeAddress())
	}

	sendBlk, err := g.qlc.Client().QGasSwap.GetPledgeSendBlock(&qlcchain.QGasPledgeParam{
		PledgeAddress: pledgeAddress,
		Amount:        qlctypes.Balance{Int: big.NewInt(params.GetAmount())},
		ToAddress:     g.ownerAddr,
	})
	if err != nil {
		g.logger.Errorf("QGas get pledge send block: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas create pledge send block: %s", sendBlk.GetHash())

	swapInfo := &types.QGasSwapInfo{
		SwapType:    types.QGasDeposit,
		State:       types.QGasPledgeInit,
		Amount:      params.GetAmount(),
		FromAddress: pledgeAddress,
		ToAddress:   g.ownerAddr,
		SendTxHash:  sendBlk.GetHash(),
		EthUserAddr: params.GetErc20ReceiverAddr(),
		StartTime:   time.Now().Unix(),
	}

	if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Errorf("insert invalid info: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas insert pledge info to %s: %s", types.QGasSwapStateToString(types.QGasPledgeInit), sendBlk.GetHash())
	return toStateBlock(sendBlk), nil
}

func (g *QGasSwapAPI) GetWithdrawBlock(ctx context.Context, param *pb.Hash) (*pb.StateBlock, error) {
	g.logger.Infof("QGas Pledge ......... (%s) ", param)
	if param == nil {
		return nil, errors.New("error params")
	}
	ethTxHash := param.GetHash()
	swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, ethTxHash, types.ETH)
	if err != nil {
		if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(ethTxHash), 0, g.cfg.EthCfg.ConfirmedHeight); err != nil {
			g.logger.Errorf("QGas withdraw eth tx not confirmed, eth[%s]", ethTxHash)
			return nil, err
		}
		amount, ethAddr, qlcAddrStr, err := g.eth.SyncBurnLog(ethTxHash)
		if err != nil {
			g.logger.Error("QGas get withdraw log: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}
		qlcAddr, err := qlctypes.HexToAddress(qlcAddrStr)
		if err != nil {
			g.logger.Error("QGas invalid address: %s, %s, eth[%s]", err, qlcAddrStr, ethTxHash)
			return nil, err
		}

		swapInfo = &types.QGasSwapInfo{
			SwapType:    types.QGasWithdraw,
			State:       types.QGasWithDrawPending,
			Amount:      amount.Int64(),
			FromAddress: g.ownerAddr,
			ToAddress:   qlcAddr,
			EthTxHash:   ethTxHash,
			EthUserAddr: ethAddr.String(),
			StartTime:   time.Now().Unix(),
		}
		if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("QGas insert invalid info: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}
		g.logger.Infof("QGas insert withdraw info to %s, eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawPending), ethTxHash)
	} else {
		if swapInfo.State == types.QGasWithDrawDone {
			g.logger.Errorf("reduplicate tx: %s, eth[%s]", err, ethTxHash)
			return nil, errors.New("reduplicate tx")
		}
	}

	if swapInfo.SendTxHash == qlctypes.ZeroHash {
		qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
		if qlcAddress == qlctypes.ZeroAddress {
			g.logger.Errorf("can not search address %s, eth[%s]", g.ownerAddr, ethTxHash)
			return nil, errors.New("can not search address")
		}
		defer g.addrPool.Enqueue(qlcAddress)

		sendBlk, err := g.qlc.Client().QGasSwap.GetWithdrawSendBlock(&qlcchain.QGasWithdrawParam{
			WithdrawAddress: swapInfo.ToAddress,
			Amount:          qlctypes.Balance{Int: big.NewInt(swapInfo.Amount)},
			FromAddress:     g.ownerAddr,
		})
		if err != nil {
			g.logger.Errorf("QGas get withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		if err := g.signQLCTx(sendBlk); err != nil {
			g.logger.Errorf("QGas sign reward block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		if err := g.qlc.ProcessAndWaitConfirmed(sendBlk); err != nil {
			g.logger.Errorf("QGas process withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		swapInfo.SendTxHash = sendBlk.GetHash()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas update withdraw send block to %s, eth[%s]", sendBlk.GetHash(), ethTxHash)
	}

	rewardBlk, err := g.qlc.Client().QGasSwap.GetWithdrawRewardBlock(swapInfo.SendTxHash)
	if err != nil {
		g.logger.Errorf("QGas get withdraw reward block: %s, eth[%s]", err, ethTxHash)
		return nil, err
	}
	return toStateBlock(rewardBlk), nil
}

func (g *QGasSwapAPI) ProcessBlock(ctx context.Context, params *pb.StateBlock) (*pb.Hash, error) {
	if params == nil {
		return nil, errors.New("nil block")
	}
	blk, err := toOriginStateBlock(params)
	if err != nil {
		g.logger.Errorf("QGas invalid block: %s, %s", err, blk.GetHash())
		return nil, err
	}

	if blk.Type == qlctypes.ContractSend { // pledge
		swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, blk.GetHash().String(), types.QLC)
		if err != nil {
			g.logger.Errorf("QGas pledge info not found: %s, qlc[%s]", err, blk.GetHash())
			return nil, err
		}

		if swapInfo.State >= types.QGasPledgeDone {
			g.logger.Errorf("QGas invalid pledge state: %s, qlc[%s]", types.QGasSwapStateToString(swapInfo.State), blk.GetHash())
			return nil, errors.New("invalid state")
		}

		if swapInfo.State == types.QGasPledgeInit {
			if !g.qlc.CheckBlockOnChain(swapInfo.SendTxHash) {
				if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
					g.logger.Errorf("QGas Process pledge send block: %s [%s]", err, blk.GetHash())
					return nil, err
				}
				g.logger.Infof("QGas process pledge send block successfully, qlc[%s]", blk.GetHash())
			}

			qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
			if qlcAddress == qlctypes.ZeroAddress {
				g.logger.Errorf("can not search address %s, qlc[%s]", g.ownerAddr, blk.GetHash())
				return nil, errors.New("can not search address")
			}
			defer g.addrPool.Enqueue(qlcAddress)

			rewardBlk, err := g.qlc.Client().QGasSwap.GetPledgeRewardBlock(swapInfo.SendTxHash)
			if err != nil {
				g.logger.Errorf("QGas get pledge reward block error: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}

			if err := g.signQLCTx(rewardBlk); err != nil {
				g.logger.Errorf("QGas sign reward block: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}

			if err := g.qlc.ProcessAndWaitConfirmed(rewardBlk); err != nil {
				g.logger.Errorf("QGas Process pledge reward block: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}
			g.logger.Infof("QGas process pledge reward block successfully: %s, qlc[%s]", rewardBlk.GetHash().String(), blk.GetHash())

			swapInfo.RewardTxHash = rewardBlk.GetHash()
			swapInfo.State = types.QGasPledgePending
			if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
				g.logger.Errorf("update invalid info: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}
			g.logger.Infof("QGas update pledge info to %s: qlc[%s]", types.QGasSwapStateToString(types.QGasPledgePending), swapInfo.SendTxHash)
			return &pb.Hash{
				Hash: rewardBlk.GetHash().String(),
			}, nil
		}

		if swapInfo.State == types.QGasPledgePending {
			return &pb.Hash{
				Hash: swapInfo.RewardTxHash.String(),
			}, nil
		}
	} else if blk.Type == qlctypes.ContractReward { // withdraw
		swapInfo, err := g.qlc.GetSwapInfoHashByWithdrawSendBlock(blk.GetLink(), g.store)
		if err != nil {
			g.logger.Errorf("QGas pledge withdraw not found: %s, %s", err, blk.GetLink())
			return nil, err
		}

		if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
			g.logger.Errorf("QGas Process withdraw reward block: %s, eth[%s]", err, swapInfo.EthTxHash)
			return nil, err
		}
		g.logger.Infof("QGas process pledge reward block successfully: %s, eth[%s]", blk.GetHash().String(), swapInfo.EthTxHash)

		swapInfo.RewardTxHash = blk.GetHash()
		swapInfo.State = types.QGasWithDrawDone
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas update withdraw info to %s: eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawDone), swapInfo.EthTxHash)
		return &pb.Hash{
			Hash: swapInfo.RewardTxHash.String(),
		}, nil
	}
	return nil, fmt.Errorf("invalid block typ: %s", blk.GetType())
}

func (g *QGasSwapAPI) GetPledgeEthOwnerSign(ctx context.Context, param *pb.Hash) (*pb.String, error) {
	g.logger.Infof("call deposit GetEthOwnerSign: %s", param)
	txHash := param.GetHash()
	if txHash == "" {
		g.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}

	swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, txHash, types.QLC)
	if err != nil {
		g.logger.Infof("qlc not locked, %s, qlc[%s]", err, txHash)
		return nil, fmt.Errorf("qlc not locked")
	}
	if swapInfo.State >= types.QGasPledgeDone {
		g.logger.Errorf("repeat operation, qlc[%s]", txHash)
		return nil, fmt.Errorf("repeat operation, qlc[%s]", txHash)
	}

	sign, err := g.signEthData(big.NewInt(swapInfo.Amount), swapInfo.EthUserAddr, hubUtil.RemoveHexPrefix(txHash))
	if err != nil {
		g.logger.Error(err)
		return nil, err
	}
	g.logger.Infof("QGas hub signed: %s. qlc[%s]", sign, txHash)
	return toString(sign), nil
}

func (g *QGasSwapAPI) signEthData(amount *big.Int, receiveAddr string, neoTxHash string) (string, error) {
	packedHash, err := packed(amount, receiveAddr, neoTxHash)
	if err != nil {
		return "", fmt.Errorf("packed: %s", err)
	}

	signature, err := g.signer.Sign(pb.SignType_ETH, g.cfg.EthCfg.OwnerAddress, packedHash)
	if err != nil {
		return "", fmt.Errorf("sign: %s", err)
	}
	sig := signature.Sign
	if len(sig) == 0 {
		return "", errors.New("invalid signature")
	}

	v := sig[len(sig)-1]
	if v == 0 || v == 1 {
		sig[len(sig)-1] = v + 27
		return hex.EncodeToString(sig), nil
	} else if v == 27 || v == 28 {
		return hex.EncodeToString(sig), nil
	} else {
		return "", fmt.Errorf("invalid signature 'v' value: %s", hex.EncodeToString(sig))
	}
}

func (g *QGasSwapAPI) signQLCTx(block *qlctypes.StateBlock) error {
	var w qlctypes.Work
	worker, err := qlctypes.NewWorker(w, block.Root())
	if err != nil {
		return err
	}
	block.Work = worker.NewWork()
	hash := block.GetHash()
	signature, err := g.signer.Sign(pb.SignType_ETH, g.cfg.EthCfg.OwnerAddress, hash.Bytes()) //todo QLC
	if err != nil {
		return fmt.Errorf("sign: %s", err)
	}
	sign, err := qlctypes.BytesToSignature(signature.GetSign())
	if err != nil {
		return fmt.Errorf("sign bytes: %s", err)
	}
	block.Signature = sign
	return nil
}
