package apis

import (
	"context"
	"fmt"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
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
	w.logger.Info("api - withdraw claim: ", request.String())
	if err := w.neo.ValidateAddress(request.GetUserNep5Addr()); err != nil {
		return nil, fmt.Errorf("invalid address: %s", request.GetUserNep5Addr())
	}
	rHash := util.Sha256(request.GetROrigin())
	if lockInfo, err := w.store.GetLockerInfo(rHash); err != nil {
		//w.logger.Errorf("get locker info: %s [%s]", err, request.GetRHash())
		return nil, err
	} else {
		if lockInfo.State != types.WithDrawNeoLockedDone {
			//w.logger.Errorf("current state is %s, [%s]", types.LockerStateToString(lockInfo.State), lockInfo.RHash)
			return nil, fmt.Errorf("invalid state: %s", types.LockerStateToString(lockInfo.State))
		}
	}

	go func() {
		lock(rHash, w.logger)
		defer unlock(rHash, w.logger)

		info, err := w.store.GetLockerInfo(rHash)
		if err != nil {
			w.logger.Error(err)
			w.store.SetLockerStateFail(info, err)
			return
		}

		tx, err := w.neo.UserUnlock(request.GetROrigin(), request.GetUserNep5Addr(), w.cfg.NEOCfg.SignerAddress)
		if err != nil {
			w.logger.Errorf("user unlock: %s, [%s]", err, rHash)
			return
		}
		w.logger.Infof("withdraw/claim neo unlock tx: %s [%s]", tx, rHash)
		info.State = types.WithDrawNeoUnLockedPending
		info.UnlockedNep5Hash = tx
		info.ROrigin = request.GetROrigin()
		info.UserAddr = request.GetUserNep5Addr()
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", rHash, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoUnLockedPending))

		var height uint32
		w.logger.Infof("waiting for neo tx %s confirmed", tx)
		if height, err = w.neo.TxVerifyAndConfirmed(tx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
			w.logger.Errorf("withdraw/txVerify(neo): %s,  %s [%s]", err, tx, rHash)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.LockedNep5Height = height
		if err = w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoLockedDone))

		tx, err = w.eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.SignerAddress)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			return
		}
		w.logger.Infof("withdraw/wrapper eth unlock: %s [%s] ", tx, rHash)
		info.State = types.WithDrawEthUnlockPending
		info.UnlockedErc20Hash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", rHash, err)
			return
		}
	}()

	return toBoolean(true), nil
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
