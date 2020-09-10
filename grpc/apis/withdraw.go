package apis

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
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

func (w *WithdrawAPI) Claim(ctx context.Context, request *pb.ClaimRequest) (*pb.String, error) {
	rHash := util.Sha256(request.GetROrigin())
	w.logger.Infof("api - withdraw claim: %s, [%s]", request.String(), rHash)
	if err := w.neo.ValidateAddress(request.GetUserNep5Addr()); err != nil {
		return nil, fmt.Errorf("invalid address: %s", request.GetUserNep5Addr())
	}

	info, err := w.store.GetLockerInfo(rHash)
	if err != nil {
		return nil, err
	}
	if info.State < types.WithDrawNeoLockedDone {
		//w.logger.Errorf("current state is %s, [%s]", types.LockerStateToString(lockInfo.State), lockInfo.RHash)
		return nil, fmt.Errorf("invalid state: %s", types.LockerStateToString(info.State))
	}

	neoUnlockTx, err := w.neo.UserUnlock(request.GetROrigin(), request.GetUserNep5Addr(), w.cfg.NEOCfg.SignerAddress)
	if err != nil {
		w.logger.Errorf("user unlock: %s, [%s]", err, rHash)
		return nil, err
	}
	w.logger.Infof("withdraw/claim neo unlock tx: %s [%s]", neoUnlockTx, rHash)
	info.State = types.WithDrawNeoUnLockedPending
	if err := w.store.UpdateLockerInfo(info); err != nil {
		return nil, err
	}
	w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoUnLockedPending))

	go func() {
		lock(rHash, w.logger)
		defer unlock(rHash, w.logger)

		info, _ := w.store.GetLockerInfo(rHash)
		if info.State >= types.WithDrawNeoUnLockedDone {
			w.logger.Infof("locker state already ahead of %s, [%s] ", types.LockerStateToString(types.WithDrawNeoUnLockedDone), info.RHash)
			w.store.SetLockerStateFail(info, err)
			return
		}

		if err := setWithDrawNeoUnLockedDone(rHash, w.neo, w.store, w.cfg.NEOCfg.ConfirmedHeight, w.logger); err != nil {
			w.logger.Errorf("set neo unlocked done: %s [%s]", err, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}

		if err := setWithDrawEthUnlockPending(rHash, w.eth, w.store, w.cfg.EthereumCfg.SignerAddress, w.logger); err != nil {
			w.logger.Errorf("set WithDrawEthUnlockPending: %s [%s]", err, info.RHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
	}()

	return toString(neoUnlockTx), nil
}
