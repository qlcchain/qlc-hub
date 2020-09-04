package apis

import (
	"context"
	"fmt"

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

type WithdrawAPI struct {
	neo    *neo.Transaction
	eth    *ethclient.Client
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client, s *store.Store) *WithdrawAPI {
	return &WithdrawAPI{
		cfg:    cfg,
		neo:    neo,
		store:  s,
		eth:    eth,
		ctx:    ctx,
		logger: log.NewLogger("api/withdraw"),
	}
}

func (w *WithdrawAPI) Unlock(ctx context.Context, request *pb.WithdrawUnlockRequest) (*pb.Boolean, error) {
	w.logger.Info("api - withdraw unlock: ", request.String())
	rHash := request.GetRHash()
	lockInfo, err := w.store.GetLockerInfo(rHash)
	if err != nil {
		w.logger.Errorf("%s: %s", rHash, err)
		return nil, err
	}
	if lockInfo.State != types.WithDrawNeoLockedDone {
		w.logger.Errorf("current [%s] is [%s]", lockInfo.RHash, types.LockerStateToString(lockInfo.State))
		return nil, fmt.Errorf("invalid state: %s", types.LockerStateToString(lockInfo.State))
	}

	go func() {
		lock(request.GetRHash(), w.logger)
		defer unlock(request.GetRHash(), w.logger)

		info, err := w.store.GetLockerInfo(request.GetRHash())
		if err != nil {
			w.logger.Error(err)
			w.store.SetLockerStateFail(info, err)
			return
		}
		if info.State >= types.WithDrawNeoUnLockedDone {
			w.logger.Infof("[%s] state already ahead [%s]", request.GetRHash(), types.LockerStateToString(types.WithDrawEthUnlockPending))
			return
		}

		height, err := w.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), w.cfg.NEOCfg.ConfirmedHeight, neo.UserUnlock)
		if err != nil {
			w.logger.Error(err)
			w.store.SetLockerStateFail(info, err)
			return
		}

		info.State = types.WithDrawNeoUnLockedDone
		info.UnlockedNep5Height = height
		info.UnlockedNep5Hash = request.GetNep5TxHash()
		info.ROrigin = request.GetROrigin()
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", request.GetRHash(), err)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoUnLockedDone))

		tx, err := eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract, w.eth)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("withdraw wrapper eth unlock: %s [%s] ", tx, rHash)
		info.State = types.WithDrawEthUnlockPending
		info.UnlockedErc20Hash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", request.GetRHash(), err)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
	}()

	return toBoolean(true), nil
}
