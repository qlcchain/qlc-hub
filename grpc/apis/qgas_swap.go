package apis

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/qlc"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type QGasSwapAPI struct {
	eth       *eth.Transaction
	bsc       *eth.Transaction
	qlc       *qlc.Transaction
	store     *gorm.DB
	cfg       *config.Config
	ctx       context.Context
	ownerAddr qlctypes.Address
	addrPool  *AddressPool
	signer    *signer.SignerClient
	logger    *zap.SugaredLogger
}

func NewQGasSwapAPI(ctx context.Context, cfg *config.Config, q *qlc.Transaction, e *eth.Transaction, b *eth.Transaction, signer *signer.SignerClient, s *gorm.DB) *QGasSwapAPI {
	api := &QGasSwapAPI{
		cfg:    cfg,
		eth:    e,
		qlc:    q,
		bsc:    b,
		ctx:    ctx,
		store:  s,
		signer: signer,
		logger: log.NewLogger("api/qgas_pledge"),
	}
	address, err := qlctypes.HexToAddress(cfg.QlcCfg.QlcOwner)
	if err != nil {
		api.logger.Fatal(err)
	}
	api.ownerAddr = address
	api.addrPool = AddressPools(address)
	go api.correctSwapState()
	return api
}

type QGasPledgeParam struct {
	PledgeAddress     qlctypes.Address
	Amount            qlctypes.Balance
	Erc20ReceiverAddr string
}

func (g *QGasSwapAPI) GetPledgeSendBlock(ctx context.Context, params *pb.QGasPledgeRequest) (*pb.StateBlockHash, error) {
	g.logger.Infof("call QGas Get Pledge Send Block......... (%s) ", params)
	if params.GetFromAddress() == "" || params.TokenMintedToAddress == "" || params.GetAmount() <= 0 {
		return nil, errors.New("error params")
	}
	chainType := types.StringToChainType(params.GetChainType())
	if chainType != types.ETH && chainType != types.BSC {
		return nil, errors.New("invalid chain")
	}

	pledgeAddress, err := qlctypes.HexToAddress(params.GetFromAddress())
	if err != nil {
		g.logger.Errorf("invalid address, %s, %s", err, params.GetFromAddress())
		return nil, err
	}

	sendBlk, err := g.qlc.Client().QGasSwap.GetPledgeSendBlock(&qlcchain.QGasPledgeParam{
		FromAddress: pledgeAddress,
		Amount:      qlctypes.Balance{Int: big.NewInt(params.GetAmount())},
		ToAddress:   g.ownerAddr,
	})
	if err != nil {
		g.logger.Errorf("QGas get pledge send block: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas create pledge send block: %s", sendBlk.GetHash())

	swapInfo := &types.QGasSwapInfo{
		SwapType:           types.QGasDeposit,
		State:              types.QGasPledgeInit,
		Chain:              chainType,
		Amount:             params.GetAmount(),
		QlcUserAddr:        pledgeAddress.String(),
		OwnerAddress:       g.ownerAddr.String(),
		QlcSendTxHash:      sendBlk.GetHash().String(),
		BlockStr:           toBlockStr(sendBlk),
		UserTxHash:         sendBlk.GetHash().String(),
		CrossChainUserAddr: params.GetTokenMintedToAddress(),
		StartTime:          time.Now().Unix(),
	}

	if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Errorf("insert invalid info: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas insert pledge info to %s, qlc[%s]", types.QGasSwapStateToString(types.QGasPledgeInit), sendBlk.GetHash())

	return &pb.StateBlockHash{
		Hash: sendBlk.GetHash().String(),
		Root: sendBlk.Root().String(),
	}, nil
}

func toBlockStr(blk *qlctypes.StateBlock) string {
	blkBytes, _ := blk.Serialize()
	return hex.EncodeToString(blkBytes)
}

func (g *QGasSwapAPI) GetWithdrawRewardBlock(ctx context.Context, param *pb.Hash) (*pb.StateBlockHash, error) {
	g.logger.Infof("call QGas Get Withdraw Block ......... (%s) ", param)
	if param == nil {
		return nil, errors.New("error params")
	}
	ethTxHash := param.GetHash()
	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, ethTxHash, types.QGasWithdraw)
	if err != nil {
		g.logger.Errorf("QGas eth tx not found", err)
		return nil, err
	} else {
		if swapInfo.State < types.QGasWithDrawPending {
			g.logger.Errorf("invalid state: %s, eth[%s]", types.QGasSwapStateToString(swapInfo.State), ethTxHash)
			return nil, errors.New("invalid  state")
		}
		if swapInfo.State == types.QGasWithDrawDone {
			g.logger.Errorf("reduplicate tx: %s, eth[%s]", err, ethTxHash)
			return nil, errors.New("reduplicate tx")
		}
	}

	if swapInfo.QlcSendTxHash == "" {
		qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
		if qlcAddress == qlctypes.ZeroAddress {
			g.logger.Errorf("can not search address %s, eth[%s]", g.ownerAddr, ethTxHash)
			return nil, errors.New("can not search address")
		}
		defer g.addrPool.Enqueue(qlcAddress)

		qlcUserAddress, err := qlctypes.HexToAddress(swapInfo.QlcUserAddr)
		if err != nil {
			g.logger.Errorf("QGas invalid address: %s, %s, eth[%s]", err, swapInfo.QlcUserAddr, ethTxHash)
			return nil, err
		}
		var linkHash qlctypes.Hash
		if err := linkHash.Of(hubUtil.RemoveHexPrefix(ethTxHash)); err != nil {
			g.logger.Errorf("QGas invalid eth tx: %s,  eth[%s]", err, ethTxHash)
			return nil, err
		}
		sendBlk, err := g.qlc.Client().QGasSwap.GetWithdrawSendBlock(&qlcchain.QGasWithdrawParam{
			ToAddress:   qlcUserAddress,
			Amount:      qlctypes.Balance{Int: big.NewInt(swapInfo.Amount)},
			FromAddress: g.ownerAddr,
			LinkHash:    linkHash,
		})
		if err != nil {
			g.logger.Errorf("QGas get withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}
		g.logger.Infof("QGas withdraw send block, %s, eth[%s]", sendBlk.GetHash(), ethTxHash)

		if err := g.signQLCTx(sendBlk); err != nil {
			g.logger.Errorf("QGas sign reward block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		if err := g.qlc.ProcessAndWaitConfirmed(sendBlk); err != nil {
			g.logger.Errorf("QGas process withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		swapInfo.QlcSendTxHash = sendBlk.GetHash().String()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas withdraw send block successfully, %s, eth[%s]", sendBlk.GetHash(), ethTxHash)
	}

	sendTxHash, err := stringToHash(swapInfo.QlcSendTxHash)
	if err != nil {
		g.logger.Errorf("QGas invalid withdraw send hash: %s, %s, eth[%s]", err, swapInfo.QlcSendTxHash, ethTxHash)
		return nil, err
	}
	rewardBlk, err := g.qlc.Client().QGasSwap.GetWithdrawRewardBlock(sendTxHash)
	if err != nil {
		g.logger.Errorf("QGas get withdraw reward block: %s, eth[%s]", err, ethTxHash)
		return nil, err
	}
	swapInfo.BlockStr = toBlockStr(rewardBlk)
	swapInfo.UserTxHash = rewardBlk.GetHash().String()
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		return nil, err
	}
	g.logger.Infof("QGas withdraw reward block, %s, eth[%s]", rewardBlk.GetHash(), ethTxHash)

	return &pb.StateBlockHash{
		Hash: rewardBlk.GetHash().String(),
		Root: rewardBlk.Root().String(),
	}, nil
}

func (g *QGasSwapAPI) convertBlock(params *pb.StateBlockSigned) (*qlctypes.StateBlock, *types.QGasSwapInfo, error) {
	signStr := params.GetSignature()
	signature, err := qlctypes.NewSignature(signStr)
	if err != nil {
		return nil, nil, fmt.Errorf("signature, %s", err)
	}

	//workStr := params.GetWork()
	//work := new(qlctypes.Work)
	//if err := work.ParseWorkHexString(workStr); err != nil {
	//	return nil, nil, fmt.Errorf("work, %s", err)
	//}

	swapInfo, err := db.GetQGasSwapInfoByUserTxHash(g.store, params.GetHash())
	blockStr := swapInfo.BlockStr
	blockBytes, err := hex.DecodeString(blockStr)
	if err != nil {
		return nil, nil, fmt.Errorf("decode string, %s", err)
	}
	block := new(qlctypes.StateBlock)
	if err := block.Deserialize(blockBytes); err != nil {
		return nil, nil, fmt.Errorf("block deserialize, %s", err)
	}
	//block.Work = *work
	block.Signature = signature
	return block, swapInfo, nil
}

func (g *QGasSwapAPI) ProcessBlock(ctx context.Context, params *pb.StateBlockSigned) (*pb.Hash, error) {
	if params == nil {
		return nil, errors.New("nil block")
	}
	g.logger.Infof("call QGas Process Block ......... (%s) ", params)
	if params.GetHash() == "" || params.GetSignature() == "" {
		g.logger.Errorf("invalid params, %s", params)
		return nil, errors.New("invalid params")
	}

	blk, swapInfo, err := g.convertBlock(params)
	if err != nil {
		g.logger.Errorf("QGas get block: %s, %s", err, params.GetHash())
		return nil, err
	}

	if blk.Type == qlctypes.ContractSend { // pledge
		g.logger.Infof("QGas pledge send block: %s", blk.GetHash().String(), types.QGasSwapStateToString(swapInfo.State))
		if blk.GetHash().String() != swapInfo.QlcSendTxHash {
			g.logger.Errorf("QGas invalid send hash: %s, qlc[%s]", swapInfo.QlcSendTxHash, blk.GetHash())
			return nil, errors.New("invalid state")
		}

		if swapInfo.State >= types.QGasPledgeDone {
			g.logger.Infof("QGas invalid pledge state: %s, qlc[%s]", types.QGasSwapStateToString(swapInfo.State), blk.GetHash())
			return &pb.Hash{
				Hash: swapInfo.QlcSendTxHash,
			}, nil
		}

		if swapInfo.State == types.QGasPledgeInit {
			if !g.qlc.CheckBlockOnChain(blk.GetHash()) {
				if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
					g.logger.Errorf("QGas Process pledge send block: %s [%s]", err, blk.GetHash())
					return nil, err
				}
				g.logger.Infof("QGas pledge send block successfully, qlc[%s]", blk.GetHash())
			}

			var receiverBlkHash qlctypes.Hash
			if receiverHash, _ := g.qlc.ReceiverBlockHash(blk.GetHash()); receiverHash.IsZero() {
				qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
				if qlcAddress == qlctypes.ZeroAddress {
					g.logger.Errorf("can not search address %s, qlc[%s]", g.ownerAddr, blk.GetHash())
					return nil, errors.New("can not search address")
				}
				defer g.addrPool.Enqueue(qlcAddress)

				rewardBlk, err := g.qlc.Client().QGasSwap.GetPledgeRewardBlock(blk.GetHash())
				if err != nil {
					g.logger.Errorf("QGas get pledge reward block error: %s, qlc[%s]", err, blk.GetHash())
					return nil, err
				}
				g.logger.Infof("QGas pledge reward block, %s, qlc[%s]", rewardBlk.GetHash().String(), blk.GetHash())
				receiverBlkHash = rewardBlk.GetHash()
				if err := g.signQLCTx(rewardBlk); err != nil {
					g.logger.Errorf("QGas sign reward block: %s, qlc[%s]", err, blk.GetHash())
					return nil, err
				}

				if err := g.qlc.ProcessAndWaitConfirmed(rewardBlk); err != nil {
					g.logger.Errorf("QGas Process pledge reward block: %s, qlc[%s]", err, blk.GetHash())
					return nil, err
				}
				g.logger.Infof("QGas pledge reward block successfully: %s, qlc[%s]", rewardBlk.GetHash().String(), blk.GetHash())
			} else {
				receiverBlkHash = receiverHash
			}
			swapInfo.QlcRewardTxHash = receiverBlkHash.String()
			swapInfo.State = types.QGasPledgePending
			if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
				g.logger.Errorf("update invalid info: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}
			g.logger.Infof("QGas update pledge info to %s: qlc[%s]", types.QGasSwapStateToString(types.QGasPledgePending), swapInfo.QlcSendTxHash)
			return &pb.Hash{
				Hash: receiverBlkHash.String(),
			}, nil
		}

		if swapInfo.State == types.QGasPledgePending {
			return &pb.Hash{
				Hash: swapInfo.QlcRewardTxHash,
			}, nil
		}
	} else if blk.Type == qlctypes.ContractReward { // withdraw
		g.logger.Infof("QGas withdraw reward block: %s", blk.String(), types.QGasSwapStateToString(swapInfo.State))
		if !g.qlc.CheckBlockOnChain(blk.GetHash()) {
			if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
				g.logger.Errorf("QGas Process withdraw reward block: %s, eth[%s]", err, swapInfo.CrossChainTxHash)
				return nil, err
			}
			g.logger.Infof("QGas process pledge reward block successfully: %s, eth[%s]", blk.GetHash().String(), swapInfo.CrossChainTxHash)

			swapInfo.QlcRewardTxHash = blk.GetHash().String()
			swapInfo.State = types.QGasWithDrawDone
			if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
				g.logger.Errorf("update invalid info: %s", err)
				return nil, err
			}
			g.logger.Infof("QGas update withdraw info to %s: eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawDone), swapInfo.CrossChainTxHash)
			g.logger.Infof("QGas withdraw successfully. eth[%s]", swapInfo.CrossChainTxHash)

		}
		return &pb.Hash{
			Hash: swapInfo.QlcRewardTxHash,
		}, nil
	}
	g.logger.Errorf("invalid block typ: %s", blk.GetType())
	return nil, fmt.Errorf("invalid block typ: %s", blk.GetType())
}

func (g *QGasSwapAPI) GetChainOwnerSign(ctx context.Context, param *pb.Hash) (*pb.String, error) {
	txHash := param.GetHash()
	if txHash == "" {
		g.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}
	g.logger.Infof("call QGas GetEthOwnerSign: %s", param)

	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, txHash, types.QGasDeposit)
	if err != nil {
		g.logger.Infof("qlc not locked, %s, qlc[%s]", err, txHash)
		return nil, fmt.Errorf("qlc not locked")
	}
	if swapInfo.State >= types.QGasPledgeDone {
		g.logger.Errorf("repeat operation, qlc[%s]", txHash)
		return nil, fmt.Errorf("repeat operation, qlc[%s]", txHash)
	}

	if swapInfo.Chain == types.ETH {
		sign, err := g.signEthData(big.NewInt(swapInfo.Amount), swapInfo.CrossChainUserAddr, hubUtil.RemoveHexPrefix(txHash), true)
		if err != nil {
			g.logger.Error(err)
			return nil, err
		}
		g.logger.Infof("QGas hub eth signed: %s. qlc[%s]", sign, txHash)
		return toString(sign), nil
	} else {
		sign, err := g.signEthData(big.NewInt(swapInfo.Amount), swapInfo.CrossChainUserAddr, hubUtil.RemoveHexPrefix(txHash), false)
		if err != nil {
			g.logger.Error(err)
			return nil, err
		}
		g.logger.Infof("QGas hub bsc signed: %s. qlc[%s]", sign, txHash)
		return toString(sign), nil
	}
}

func (g *QGasSwapAPI) PledgeChainTxSent(ctx context.Context, param *pb.EthTxSentRequest) (*pb.Boolean, error) {
	if param == nil {
		return nil, errors.New("nil param")
	}
	g.logger.Infof("call QGas pledge EthTransactionSent: %s", param)
	qlcTxHash := param.GetQlcTxHash()
	ethTxHash := param.GetChainTxHash()

	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, qlcTxHash, types.QGasDeposit)
	if err != nil {
		g.logger.Errorf("get swap info error: %s, qlc[%s]", err, qlcTxHash)
		return nil, errors.New("swap info not found")
	}
	if swapInfo.State >= types.QGasPledgeDone {
		return toBoolean(true), nil
	}
	if swapInfo.State < types.QGasPledgePending {
		g.logger.Errorf("invalid pledge state: %s", types.QGasSwapStateToString(swapInfo.State))
		return nil, errors.New("invalid state")
	}
	go func() {
		swapInfo.CrossChainTxHash = ethTxHash
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Error(err)
			return
		}
		g.processPledgeEthTx(swapInfo)
	}()
	return toBoolean(true), nil
}

func (g *QGasSwapAPI) processPledgeEthTx(swapInfo *types.QGasSwapInfo) {
	crossChainTxHash := swapInfo.CrossChainTxHash
	qlcTxHash := swapInfo.QlcSendTxHash

	var amount *big.Int
	var crossChainAddress common.Address
	var qlcTx string
	var err error
	if swapInfo.Chain == types.ETH {
		if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(crossChainTxHash), 0, g.cfg.EthCfg.EthConfirmedHeight); err != nil {
			g.logger.Errorf("QGas pledge eth tx confirmed: %s", err)
			return
		}
		g.logger.Infof("QGas pledge eth Tx confirmed, %s, qlc[%s]", crossChainTxHash, qlcTxHash)

		if amount, crossChainAddress, qlcTx, err = g.eth.QGasSyncMintLog(crossChainTxHash); err != nil {
			g.logger.Errorf("mint log: %s", err)
			return
		}
	} else {
		if err := g.bsc.WaitTxVerifyAndConfirmed(common.HexToHash(crossChainTxHash), 0, g.cfg.BscCfg.BscConfirmedHeight); err != nil {
			g.logger.Errorf("QGas pledge eth tx confirmed: %s", err)
			return
		}
		g.logger.Infof("QGas pledge eth Tx confirmed, %s, qlc[%s]", crossChainTxHash, qlcTxHash)

		if amount, crossChainAddress, qlcTx, err = g.bsc.QGasSyncMintLog(crossChainTxHash); err != nil {
			g.logger.Errorf("mint log: %s", err)
			return
		}
	}

	if amount.Int64() != swapInfo.Amount || strings.ToLower(crossChainAddress.String()) != strings.ToLower(swapInfo.CrossChainUserAddr) || qlcTx != qlcTxHash {
		g.logger.Errorf("swap info not match: %s, amount %d, address %s", qlcTx, amount.Int64(), crossChainAddress.String())
		return
	}

	swapInfo.State = types.QGasPledgeDone
	swapInfo.CrossChainTxHash = crossChainTxHash
	swapInfo.CrossChainUserAddr = crossChainAddress.String()
	swapInfo.Amount = amount.Int64()
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Error(err)
		return
	}
	g.logger.Infof("update state to %s, qlc[%s]", types.QGasSwapStateToString(types.QGasPledgeDone), qlcTx)
	g.logger.Infof("QGas pledge successfully. qlc[%s]", qlcTx)
}

func (g *QGasSwapAPI) WithdrawChainTxSent(ctx context.Context, param *pb.QGasWithdrawRequest) (*pb.Boolean, error) {
	if param == nil {
		return nil, errors.New("nil param")
	}
	g.logger.Infof("call QGas withdraw EthTransactionSent: %s", param)
	crossChainTxHash := param.GetHash()
	chainType := types.StringToChainType(param.GetChainType())
	if chainType != types.ETH && chainType != types.BSC {
		return nil, errors.New("invalid chain")
	}

	_, err := db.GetQGasSwapInfoByUniqueID(g.store, crossChainTxHash, types.QGasWithdraw)
	if err == nil {
		return toBoolean(true), nil
	}

	go func() {
		swapInfo := &types.QGasSwapInfo{
			SwapType:         types.QGasWithdraw,
			State:            types.QGasWithDrawInit,
			Chain:            chainType,
			CrossChainTxHash: crossChainTxHash,
			OwnerAddress:     g.ownerAddr.String(),
			StartTime:        time.Now().Unix(),
		}
		if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("QGas insert invalid info: %s, eth[%s]", err, crossChainTxHash)
			return
		}
		g.logger.Infof("QGas insert withdraw info to %s, eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawInit), crossChainTxHash)

		if err := g.processWithdrawChainTx(swapInfo); err != nil {
			g.logger.Errorf("QGas eth tx confirmed:  %s, eth[%s]", err, crossChainTxHash)
		}
	}()
	return toBoolean(true), nil
}

func (g *QGasSwapAPI) processWithdrawChainTx(swapInfo *types.QGasSwapInfo) error {
	crossChainTxHash := swapInfo.CrossChainTxHash

	var amount *big.Int
	var user common.Address
	var qlcAddrStr string
	var err error
	if swapInfo.Chain == types.ETH {
		if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(crossChainTxHash), 0, g.cfg.EthCfg.EthConfirmedHeight); err != nil {
			return fmt.Errorf("tx confirmed: %s", err)
		}
		g.logger.Infof("QGas withdraw eth tx confirmed, eth[%s]", crossChainTxHash)
		if amount, user, qlcAddrStr, err = g.eth.QGasSyncBurnLog(crossChainTxHash); err != nil {
			return fmt.Errorf("get burn log, %s", err)
		}
	} else {
		if err := g.bsc.WaitTxVerifyAndConfirmed(common.HexToHash(crossChainTxHash), 0, g.cfg.BscCfg.BscConfirmedHeight); err != nil {
			return fmt.Errorf("tx confirmed: %s", err)
		}
		g.logger.Infof("QGas withdraw eth tx confirmed, eth[%s]", crossChainTxHash)
		if amount, user, qlcAddrStr, err = g.bsc.QGasSyncBurnLog(crossChainTxHash); err != nil {
			return fmt.Errorf("get burn log, %s", err)
		}
	}
	qlcAddr, err := qlctypes.HexToAddress(qlcAddrStr)
	if err != nil {
		return err
	}
	swapInfo.Amount = amount.Int64()
	swapInfo.QlcUserAddr = qlcAddr.String()
	swapInfo.CrossChainUserAddr = user.String()
	swapInfo.State = types.QGasWithDrawPending
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		return err
	}
	g.logger.Infof("QGas update withdraw info to %s, eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawPending), crossChainTxHash)
	return nil
}

func (g *QGasSwapAPI) signEthData(amount *big.Int, receiveAddr string, neoTxHash string, isEth bool) (string, error) {
	var sig []byte
	if isEth {
		packedHash, err := packedWithIndex(amount, receiveAddr, neoTxHash, big.NewInt(0))
		if err != nil {
			return "", fmt.Errorf("packed: %s", err)
		}

		signature, err := g.signer.Sign(pb.SignType_ETH, g.cfg.EthCfg.EthQGasOwner, packedHash)
		if err != nil {
			return "", fmt.Errorf("sign: %s", err)
		}
		sig = signature.Sign
	} else {
		packedHash, err := packedWithIndex(amount, receiveAddr, neoTxHash, big.NewInt(1))
		if err != nil {
			return "", fmt.Errorf("packed: %s", err)
		}

		signature, err := g.signer.Sign(pb.SignType_BSC, g.cfg.BscCfg.BscQGasOwner, packedHash)
		if err != nil {
			return "", fmt.Errorf("sign: %s", err)
		}
		sig = signature.Sign
	}
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
	//var w qlctypes.Work
	//worker, err := qlctypes.NewWorker(w, block.Root())
	//if err != nil {
	//	return err
	//}
	//block.Work = worker.NewWork()
	hash := block.GetHash()
	signature, err := g.signer.Sign(pb.SignType_QLC, g.cfg.QlcCfg.QlcOwner, hash.Bytes())
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

func stringToHash(h string) (qlctypes.Hash, error) {
	var hash qlctypes.Hash
	if err := hash.Of(h); err != nil {
		return qlctypes.ZeroHash, err
	}
	return hash, nil
}

func toQGasSwapInfo(info *types.QGasSwapInfo) *pb.QGasSwapInfo {
	return &pb.QGasSwapInfo{
		State:           int32(info.State),
		StateStr:        types.QGasSwapStateToString(info.State),
		Chain:           types.ChainTypeToString(info.Chain),
		Amount:          info.Amount,
		ChainTxHash:     info.CrossChainTxHash,
		QlcSendTxHash:   info.QlcSendTxHash,
		QlcRewardTxHash: info.QlcRewardTxHash,
		ChainUserAddr:   info.CrossChainUserAddr,
		QlcUserAddr:     info.QlcUserAddr,
		StartTime:       time.Unix(info.StartTime, 0).Format(time.RFC3339),
		LastModifyTime:  time.Unix(info.LastModifyTime, 0).Format(time.RFC3339),
	}
}

func toQGasSwapInfos(infos []*types.QGasSwapInfo) *pb.QGasSwapInfos {
	r := make([]*pb.QGasSwapInfo, 0)
	for _, info := range infos {
		r = append(r, toQGasSwapInfo(info))
	}
	return &pb.QGasSwapInfos{
		Infos: r,
	}
}

func (g *QGasSwapAPI) SwapInfoByTxHash(ctx context.Context, h *pb.Hash) (*pb.QGasSwapInfo, error) {
	hash := h.GetHash()
	if !(len(hash) == 66 || len(hash) == 64) {
		return nil, fmt.Errorf("invalid hash: %s", hash)
	}
	info, err := db.GetQGasSwapInfoByUniqueID(g.store, hash, types.QGasWithdraw)
	if err != nil {
		info, err := db.GetQGasSwapInfoByUniqueID(g.store, hash, types.QGasDeposit)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		} else {
			return toQGasSwapInfo(info), nil
		}
	} else {
		return toQGasSwapInfo(info), nil
	}
}

func (g *QGasSwapAPI) SwapInfoList(ctx context.Context, offset *pb.Offset) (*pb.QGasSwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetPage(), offset.GetPageSize())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()

	infos, err := db.GetQGasSwapInfos(g.store, offset.GetChain(), int(page), int(pageSize))
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toQGasSwapInfos(infos), nil
}

func (g *QGasSwapAPI) SwapInfosByAddress(ctx context.Context, offset *pb.AddrAndOffset) (*pb.QGasSwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetAddress())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	addr := offset.GetAddress()

	if err := g.qlc.ValidateAddress(addr); err == nil {
		infos, err := db.GetQGasSwapInfosByUserAddr(g.store, int(page), int(pageSize), addr, offset.GetChain(), false)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toQGasSwapInfos(infos), nil
	} else {
		infos, err := db.GetQGasSwapInfosByUserAddr(g.store, int(page), int(pageSize), addr, offset.GetChain(), true)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toQGasSwapInfos(infos), nil
	}
}

func (g *QGasSwapAPI) SwapInfosByState(ctx context.Context, offset *pb.StateAndOffset) (*pb.QGasSwapInfos, error) {
	if types.StringToQGasSwapState(offset.GetState()) == types.QGasInvalid {
		return nil, fmt.Errorf("invalid state: %s", offset.GetState())
	}
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetState())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	state := types.StringToQGasSwapState(offset.GetState())
	infos, err := db.GetQGasSwapInfosByState(g.store, int(page), int(pageSize), state)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toQGasSwapInfos(infos), nil
}

func (g *QGasSwapAPI) SwapInfosCount(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	count := make(map[string]int64)
	infos, err := db.GetQGasSwapInfos(g.store, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.QGasPledgeDone {
			count["QGasPledgeTotal"] = count["QGasPledgeTotal"] + 1
		} else {
			count["QGasWithdrawTotal"] = count["QGasWithdrawTotal"] + 1
		}
		count[types.QGasSwapStateToString(info.State)] = count[types.QGasSwapStateToString(info.State)] + 1
	}
	return &pb.Map{
		Count: count,
	}, nil
}

func (g *QGasSwapAPI) SwapInfosAmount(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	amount := make(map[string]int64)
	infos, err := db.GetQGasSwapInfos(g.store, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.QGasPledgeDone {
			amount["QGasPledgeTotal"] = amount["QGasPledgeTotal"] + info.Amount
		} else {
			amount["QGasWithdrawTotal"] = amount["QGasWithdrawTotal"] + info.Amount
		}
		amount[types.QGasSwapStateToString(info.State)] = amount[types.QGasSwapStateToString(info.State)] + info.Amount
	}
	return &pb.Map{
		Count: amount,
	}, nil
}

// update by state
func (g *QGasSwapAPI) correctSwapState() error {
	vTicker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-vTicker.C:
			infos, err := db.GetQGasSwapInfos(g.store, "", 0, 0)
			if err != nil {
				g.logger.Error(err)
				continue
			}
			for _, info := range infos {
				if info.State == types.QGasPledgePending && time.Now().Unix()-info.LastModifyTime > 60*6 {
					if info.CrossChainTxHash != "" {
						g.processPledgeEthTx(info)
					} else {
						var amount *big.Int
						var err error
						if info.Chain == types.ETH {
							amount, err = g.eth.GetQGasLockedAmountByQLCTxHash(info.QlcSendTxHash)
						} else {
							amount, err = g.bsc.GetQGasLockedAmountByQLCTxHash(info.QlcSendTxHash)
						}
						if err == nil && amount.Int64() == info.Amount {
							info.State = types.QGasPledgeDone //can not get tx hash in eth contract
							if err := db.UpdateQGasSwapInfo(g.store, info); err == nil {
								g.logger.Infof("correct qgas deposit swap state: qlc[%s]", info.QlcSendTxHash)
							}
						}
					}
				}
				if info.State == types.QGasWithDrawInit && time.Now().Unix()-info.LastModifyTime > 60*5 {
					if err := g.processWithdrawChainTx(info); err != nil {
						g.logger.Errorf("QGas eth tx confirmed:  %s, eth[%s]", err, info.CrossChainTxHash)
					}
				}
				if info.State == types.QGasWithDrawPending {

				}
			}
		}
	}
}

func packedWithIndex(amount *big.Int, receiveAddr string, neoTxHash string, index *big.Int) ([]byte, error) {
	packedBytes := make([]byte, 0)

	amountBytes := bytes.Repeat([]byte{0}, 32)
	aBytes := amount.Bytes()
	copy(amountBytes[len(amountBytes)-len(aBytes):], aBytes)
	packedBytes = append(packedBytes, amountBytes...)

	addr := common.HexToAddress(receiveAddr)
	packedBytes = append(packedBytes, addr.Bytes()...)

	nHash := common.HexToHash(neoTxHash)
	packedBytes = append(packedBytes, nHash.Bytes()...)

	indexBytes := bytes.Repeat([]byte{0}, 32)
	iBytes := index.Bytes()
	copy(indexBytes[len(indexBytes)-len(iBytes):], iBytes)
	packedBytes = append(packedBytes, indexBytes...)

	hash := sha256.Sum256(packedBytes)

	return hash[:], nil
}
