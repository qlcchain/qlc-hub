package apis

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"go.uber.org/zap"
)

type EthAPI struct {
	contractAddr string
	ethClient    *ethclient.Client
	cfg          *config.Config
	ctx          context.Context
	logger       *zap.SugaredLogger
}

func NewEthAPI(ctx context.Context, cfg *config.Config) (*EthAPI, error) {
	//todo check address validity

	ethClient, err := ethclient.Dial(cfg.EthereumCfg.EndPoint)
	if err != nil {
		return nil, fmt.Errorf("eth client dail: %s", err)
	}
	api := &EthAPI{
		cfg:          cfg,
		contractAddr: cfg.EthereumCfg.Contract,
		ethClient:    ethClient,
		ctx:          ctx,
		logger:       log.NewLogger("api/withdraw"),
	}
	return api, nil
}

func (w *EthAPI) DepositLock(ctx context.Context, request *pb.DepositLockRequest) (*pb.Boolean, error) {
	//todo check params

	go func() {
		instance, opts, err := eth.GetTransactor(w.ethClient, w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract)
		if err != nil {
			w.logger.Error(err)
			return
		}
		bigAmount := big.NewInt(request.Amount)
		rHashBytes, err := util.HexStringToBytes32(request.RHash)
		if err != nil {
			w.logger.Error(err)
			return
		}
		tx, err := instance.IssueLock(opts, rHashBytes, bigAmount)
		if err != nil {
			w.logger.Error(err)
			return
		}
		w.logger.Info("tx hash: ", tx.Hash().Hex())
	}()
	//todo add state
	return toBoolean(true), nil
}

func (w *EthAPI) WithdrawUnlock(ctx context.Context, request *pb.WithdrawUnlockRequest) (*pb.Boolean, error) {
	//todo check params

	go func() {
		instance, opts, err := eth.GetTransactor(w.ethClient, w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract)
		if err != nil {
			w.logger.Error(err)
			return
		}

		rOriginBytes, err := util.HexStringToBytes32(request.ROrigin)
		if err != nil {
			w.logger.Error(err)
			return
		}
		rHashBytes, err := util.HexStringToBytes32(sha256(request.ROrigin))
		if err != nil {
			w.logger.Error(err)
			return
		}
		tx, err := instance.DestoryUnlock(opts, rHashBytes, rOriginBytes)
		if err != nil {
			w.logger.Error(err)
			return
		}
		w.logger.Info("tx hash: ", tx.Hash().Hex())
		//todo add state
	}()

	return toBoolean(true), nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}
