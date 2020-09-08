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
		info.UnlockedNeoHash = tx
		info.ROrigin = request.GetROrigin()
		info.NeoUserAddr = request.GetUserNep5Addr()
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", rHash, err)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoUnLockedPending))

		var height uint32
		w.logger.Infof("waiting for neo tx %s confirmed", tx)
		if height, err = w.neo.TxVerifyAndConfirmed(tx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
			w.logger.Errorf("withdraw/txVerify(neo): %s,  %s [%s]", err, tx, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.UnlockedNeoHeight = height
		if err = w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoLockedDone))

		tx, err = w.eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.SignerAddress)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("withdraw/wrapper eth unlock: %s [%s] ", tx, rHash)

		info.State = types.WithDrawEthUnlockPending
		info.UnlockedEthHash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.logger.Errorf("%s: %s", rHash, err)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
	}()

	return toBoolean(true), nil
}
