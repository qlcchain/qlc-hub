package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

type DepositAPI struct {
	contractAddr   string
	ethClient      *ethclient.Client
	neoTransaction *neo.Transaction
	store          *store.Store
	cfg            *config.Config
	ctx            context.Context
	logger         *zap.SugaredLogger
}

func NewDepositAPI(ctx context.Context, cfg *config.Config) (*DepositAPI, error) {
	//todo check address validity

	//todo close client
	ethClient, err := ethclient.Dial(cfg.EthereumCfg.EndPoint)
	if err != nil {
		return nil, fmt.Errorf("eth client dail: %s", err)
	}
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	nt, err := neo.NewTransaction(cfg.NEOCfg.EndPoint, cfg.NEOCfg.Contract)
	if err != nil {
		return nil, fmt.Errorf("neo transaction: %s", err)
	}
	api := &DepositAPI{
		cfg:            cfg,
		contractAddr:   cfg.EthereumCfg.Contract,
		neoTransaction: nt,
		ethClient:      ethClient,
		ctx:            ctx,
		store:          store,
		logger:         log.NewLogger("api/deposit"),
	}
	return api, nil
}

func (w *DepositAPI) Lock(ctx context.Context, request *pb.DepositLockRequest) (*pb.Boolean, error) {
	w.logger.Info("deposit lock: ", request.String())
	//todo check params

	// init info
	info := &types.LockerInfo{
		State:          types.DepositInit,
		RHash:          request.GetRHash(),
		Amount:         request.GetAmount(),
		LockedNep5Hash: request.GetNep5TxHash(),
	}
	if err := w.store.AddLockerInfo(info); err != nil {
		return nil, err
	}
	w.logger.Infof("add [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositInit))
	go func() {
		b, height, err := neo.TxVerifyAndConfirmed(request.GetNep5TxHash(), neoConfirmedHeight, w.neoTransaction)
		if !b || err != nil {
			w.logger.Errorf("neo tx confirmed: %s, %v [%s]", err, b, request.GetRHash())
			return
		}

		// init info
		info.State = types.DepositNeoLockedDone
		info.LockedNep5Height = height
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoLockedDone))
		// wrapper to eth lock
		tx, err := eth.WrapperLock(request.GetRHash(), w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract, request.GetAmount(), w.ethClient)
		if err != nil {
			w.logger.Error(err)
			return
		}
		w.logger.Infof("deposit/wrapper eth lock: %s [%s]", request.GetRHash(), tx)
		info.State = types.DepositEthLockedPending
		info.LockedErc20Hash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedPending))
	}()
	return toBoolean(true), nil
}

func (w *DepositAPI) FetchNotice(ctx context.Context, request *pb.FetchNoticeRequest) (*pb.Boolean, error) {
	panic("implement me")
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}