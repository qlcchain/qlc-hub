package apis

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
)

type DepositAPI struct {
	contractAddr string
	eth          *ethclient.Client
	neo          *neo.Transaction
	store        *store.Store
	cfg          *config.Config
	ctx          context.Context
	logger       *zap.SugaredLogger
}

func NewDepositAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client) (*DepositAPI, error) {
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	api := &DepositAPI{
		cfg:          cfg,
		contractAddr: cfg.EthereumCfg.Contract,
		neo:          neo,
		eth:          eth,
		ctx:          ctx,
		store:        store,
		logger:       log.NewLogger("api/deposit"),
	}
	return api, nil
}

func (w *DepositAPI) Lock(ctx context.Context, request *pb.DepositLockRequest) (*pb.Boolean, error) {
	w.logger.Info("api - deposit lock: ", request.String())
	//todo check params

	// init info
	info := &types.LockerInfo{
		State:          types.DepositInit,
		RHash:          request.GetRHash(),
		LockedNep5Hash: request.GetNep5TxHash(),
	}
	if err := w.store.AddLockerInfo(info); err != nil {
		return nil, err
	}
	w.logger.Infof("[%s] add state to [%s]", info.RHash, types.LockerStateToString(types.DepositInit))
	go func() {
		b, height, err := w.neo.TxVerifyAndConfirmed(request.GetNep5TxHash(), neoConfirmedHeight)
		if !b || err != nil {
			w.logger.Errorf("neo tx confirmed: %s, %v [%s]", err, b, request.GetRHash())
			return
		}

		swapInfo, err := w.neo.QuerySwapInfo(request.GetRHash())
		if err != nil {
			w.logger.Errorf("query swap info: %s", err)
		}

		// init info
		info.State = types.DepositNeoLockedDone
		info.LockedNep5Height = height
		info.Amount = swapInfo.Amount
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoLockedDone))

		// wrapper to eth lock
		tx, err := eth.WrapperLock(request.GetRHash(), w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract, swapInfo.Amount, w.eth)
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
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedPending))
	}()
	return toBoolean(true), nil
}

func (w *DepositAPI) FetchNotice(ctx context.Context, request *pb.FetchNoticeRequest) (*pb.Boolean, error) {
	w.logger.Info("api - deposit fetchNotice: ", request.String())
	//todo param verify

	info, err := w.store.GetLockerInfo(request.GetRHash())
	if err != nil {
		return nil, err
	}
	go func() {
		b, height, err := w.neo.TxVerifyAndConfirmed(request.GetNep5TxHash(), neoConfirmedHeight)
		if !b || err != nil {
			w.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, request.GetRHash())
			return
		}
		info.State = types.DepositNeoFetchDone
		info.UnlockedNep5Height = height
		info.UnlockedNep5Hash = request.GetNep5TxHash()
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoFetchDone))
	}()
	return toBoolean(true), nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}
