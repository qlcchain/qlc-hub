package apis

import (
	"context"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"go.uber.org/zap"
)

type WithdrawAPI struct {
	neo    *neo.Transaction
	eth    *eth.Transaction
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, s *store.Store) *WithdrawAPI {
	return &WithdrawAPI{
		cfg:    cfg,
		neo:    neo,
		store:  s,
		eth:    eth,
		ctx:    ctx,
		logger: log.NewLogger("api/withdraw"),
	}
}

func (w *WithdrawAPI) Claim(ctx context.Context, request *pb.ClaimRequest) (*pb.Boolean, error) {
	panic("implement me")
}

//
//func (w *WithdrawAPI) Unlock(ctx context.Context, request *pb.WithdrawUnlockRequest) (*pb.Boolean, error) {
//	w.logger.Info("api - withdraw unlock: ", request.String())
//	lockInfo, err := w.store.GetLockerInfo(request.GetRHash())
//	if err != nil {
//		//w.logger.Errorf("get locker info: %s [%s]", err, request.GetRHash())
//		return nil, err
//	}
//	if lockInfo.State != types.WithDrawNeoLockedDone {
//		//w.logger.Errorf("current state is %s, [%s]", types.LockerStateToString(lockInfo.State), lockInfo.RHash)
//		return nil, fmt.Errorf("invalid state: %s", types.LockerStateToString(lockInfo.State))
//	}
//
//	go func() {
//		lock(request.GetRHash(), w.logger)
//		defer unlock(request.GetRHash(), w.logger)
//
//		info, err := w.store.GetLockerInfo(request.GetRHash())
//		if err != nil {
//			w.logger.Error(err)
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//		if info.State >= types.WithDrawEthUnlockPending {
//			w.logger.Infof("[%s] state already ahead [%s]", request.GetRHash(), types.LockerStateToString(types.WithDrawEthUnlockPending))
//			return
//		}
//
//		w.logger.Infof("check nep5 tx %s [%s]", request.GetNep5TxHash(), request.GetRHash())
//		height, err := w.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), w.cfg.NEOCfg.ConfirmedHeight, neo.UserUnlock)
//		if err != nil {
//			w.logger.Error(err)
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//
//		swapInfo, err := w.neo.QuerySwapInfo(request.GetRHash())
//		if err != nil {
//			w.logger.Errorf("query swap info: %s", err)
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//		w.logger.Infof("swap info: %s", util.ToString(swapInfo))
//
//		info.State = types.WithDrawNeoUnLockedDone
//		info.UserAddr = swapInfo.UserNep5Address
//		info.UnlockedNep5Height = height
//		info.UnlockedNep5Hash = request.GetNep5TxHash()
//		info.ROrigin = swapInfo.OriginText
//		if err := w.store.UpdateLockerInfo(info); err != nil {
//			w.logger.Errorf("%s: %s", request.GetRHash(), err)
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoUnLockedDone))
//
//		tx, err := eth.WrapperUnlock(request.GetRHash(), request.GetROrigin(), w.cfg.EthereumCfg.Address, w.cfg.EthereumCfg.Contract, w.eth)
//		if err != nil {
//			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, request.GetRHash())
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//		w.logger.Infof("withdraw wrapper eth unlock: %s [%s] ", tx, request.GetRHash())
//		info.State = types.WithDrawEthUnlockPending
//		info.UnlockedErc20Hash = tx
//		if err := w.store.UpdateLockerInfo(info); err != nil {
//			w.logger.Errorf("%s: %s", request.GetRHash(), err)
//			w.store.SetLockerStateFail(info, err)
//			return
//		}
//		w.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
//	}()
//
//	return toBoolean(true), nil
//}
