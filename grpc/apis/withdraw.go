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
			w.logger.Infof("[%s] state [%s] already ahead [%s]", info.RHash, types.LockerStateToString(info.State), types.LockerStateToString(types.WithDrawNeoUnLockedDone))
			return
		}

		swapInfo, err := w.neo.QuerySwapInfoAndConfirmedTx(rHash, neo.UserUnlock, w.cfg.NEOCfg.ConfirmedHeight)
		if err != nil {
			w.logger.Infof("query swap info: %s", err, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
		info.State = types.WithDrawNeoUnLockedDone
		info.UnlockedNeoHash = swapInfo.TxIdOut
		info.ROrigin = swapInfo.OriginText
		info.NeoUserAddr = swapInfo.UserNeoAddress
		info.UnlockedNeoHeight = swapInfo.UnlockedHeight

		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawNeoUnLockedDone))

		ethUnlockTx, err := w.eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.SignerAddress)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("withdraw/wrapper eth unlock: %s [%s] ", ethUnlockTx, rHash)

		info.State = types.WithDrawEthUnlockPending
		if err := w.store.UpdateLockerInfo(info); err != nil {
			w.store.SetLockerStateFail(info, err)
			return
		}
		w.logger.Infof("set [%s] state to [%s]", rHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
	}()

	return toString(neoUnlockTx), nil
}
