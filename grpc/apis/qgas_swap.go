package apis

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/qlc"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type QGasSwapAPI struct {
	eth       *eth.Transaction
	qlc       *qlc.Transaction
	store     *gorm.DB
	cfg       *config.Config
	ctx       context.Context
	ownerAddr qlctypes.Address
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
	return api
}

type QGasPledgeParam struct {
	PledgeAddress qlctypes.Address
	Amount        qlctypes.Balance
}

func (g *QGasSwapAPI) Pledge(ctx context.Context, params *pb.QGasPledgeRequest) (*pb.StateBlock, error) {
	g.logger.Infof("QGas Pledge ......... (%s) ", params)
	if params.GetPledgeAddress() == "" || params.GetAmount() <= 0 {
		return nil, errors.New("error params")
	}
	pledgeAddress, err := qlctypes.HexToAddress(params.GetPledgeAddress())
	if err != nil {
		return nil, fmt.Errorf("invalid address, %s", params.GetPledgeAddress())
	}

	sendBlk, err := g.qlc.Client().QGasSwap.GetPledgeSendBlock(&qlcchain.QGasPledgeInfo{
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
		Amount:      params.Amount,
		FromAddress: pledgeAddress,
		ToAddress:   g.ownerAddr,
		SendTxHash:  sendBlk.GetHash(),
		StartTime:   time.Now().Unix(),
	}

	if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Errorf("insert invalid info: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas insert pledge info to %s: %s", types.QGasSwapStateToString(types.QGasPledgeInit), sendBlk.GetHash())
	return toStateBlock(sendBlk), nil
}

func (g *QGasSwapAPI) Withdraw(ctx context.Context, param *pb.Hash) (*pb.StateBlock, error) {
	g.logger.Infof("QGas Pledge ......... (%s) ", param)
	if param == nil {
		return nil, errors.New("error params")
	}
	ethTxHash := param.GetHash()
	swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, ethTxHash, types.ETH)
	if swapInfo == nil {
		if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(ethTxHash), 0, 1); err != nil {
			g.logger.Errorf("QGas withdraw eth tx not confirmed %s", ethTxHash)
			return nil, err
		}
		amount, ethAddr, qlcAddrStr, err := g.eth.SyncBurnLog(ethTxHash)
		if err != nil {
			g.logger.Error("QGas get withdraw log: %s", err)
			return nil, err
		}
		qlcAddr, err := qlctypes.HexToAddress(qlcAddrStr)
		if err != nil {
			g.logger.Error("QGas invalid address: %s, %s", err, qlcAddrStr)
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
			g.logger.Errorf("QGas insert invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas insert withdraw info to %s: %s", types.QGasSwapStateToString(types.QGasWithDrawPending), ethTxHash)
	} else {
		if swapInfo.State == types.QGasWithDrawDone {
			return nil, fmt.Errorf("reduplicate tx: %s", err)
		}
	}

	if swapInfo.SendTxHash == qlctypes.ZeroHash {
		sendBlk, err := g.qlc.Client().QGasSwap.GetWithdrawSendBlock(&qlcchain.QGasWithdrawInfo{
			WithdrawAddress: swapInfo.ToAddress,
			Amount:          qlctypes.Balance{Int: big.NewInt(swapInfo.Amount)},
			FromAddress:     g.ownerAddr,
		})
		if err != nil {
			g.logger.Errorf("QGas get withdraw send block: %s", err)
			return nil, err
		}

		if _, err := g.qlc.Client().Ledger.Process(sendBlk); err != nil {
			g.logger.Errorf("QGas process withdraw send block: %s", err)
			return nil, err
		}
		swapInfo.SendTxHash = sendBlk.GetHash()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas update withdraw send block to %s: %s", sendBlk.GetHash(), ethTxHash)
	}

	rewardBlk, err := g.qlc.Client().QGasSwap.GetWithdrawRewardBlock(swapInfo.SendTxHash)
	if err != nil {
		g.logger.Errorf("QGas get withdraw reward block: %s", err)
		return nil, err
	}
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Errorf("update invalid info: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas update withdraw reward block to %s: %s", rewardBlk.GetHash(), ethTxHash)
	return toStateBlock(rewardBlk), nil
}

func (g *QGasSwapAPI) ProcessBlock(ctx context.Context, params *pb.StateBlock) (*pb.Hash, error) {
	if params == nil {
		return nil, errors.New("nil block")
	}
	blk, err := toOriginStateBlock(params)
	if err != nil {
		g.logger.Errorf("QGas invalid block: %s", err)
		return nil, err
	}

	if blk.Type == qlctypes.ContractSend {
		swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, blk.GetHash().String(), types.QLC)
		if err != nil {
			g.logger.Errorf("QGas pledge info not found: %s", err)
			return nil, err
		}
		if swapInfo.State > types.QGasPledgePending {
			g.logger.Errorf("QGas invalid pledge state: %s", types.QGasSwapStateToString(swapInfo.State))
			return nil, errors.New("invalid state")
		}
		h, err := g.qlc.Client().Ledger.Process(blk)
		if err != nil {
			g.logger.Errorf("QGas Process pledge send block: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas process pledge send block successfully: %s", h.String())

		swapInfo.State = types.QGasPledgePending
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas update pledge info to %s: %s", types.QGasSwapStateToString(types.QGasPledgePending), h)
		go func() {

		}()
		return &pb.Hash{
			Hash: h.String(),
		}, nil
	} else if blk.Type == qlctypes.ContractReward {
		swapInfo, err := db.GetQGasSwapInfoByTxHash(g.store, blk.GetHash().String(), types.ETH)
		if err != nil {
			g.logger.Errorf("QGas pledge withdraw not found: %s", err)
			return nil, err
		}

		if _, err := g.qlc.Client().Ledger.Process(blk); err != nil {
			g.logger.Errorf("QGas process withdraw reward block: %s", err)
			return nil, err
		}
		swapInfo.RewardTxHash = blk.GetHash()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas update withdraw info to %s: %s", types.QGasSwapStateToString(types.QGasWithDrawDone), swapInfo.EthTxHash)
	}
	return nil, fmt.Errorf("invalid block typ: %s", blk.GetType())
}
