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

type DepositAPI struct {
	eth    *eth.Transaction
	neo    *neo.Transaction
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewDepositAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, s *store.Store) *DepositAPI {
	return &DepositAPI{
		cfg:    cfg,
		neo:    neo,
		eth:    eth,
		ctx:    ctx,
		store:  s,
		logger: log.NewLogger("api/deposit"),
	}
}

func (d *DepositAPI) Lock(ctx context.Context, request *pb.DepositLockRequest) (*pb.Boolean, error) {
	d.logger.Info("api - deposit lock: ", request.String())
	if err := d.checkLockParams(request); err != nil {
		d.logger.Error(err)
		return nil, err
	}

	if lockerInfo, err := d.store.GetLockerInfo(request.GetRHash()); err == nil {
		if lockerInfo.State == types.DepositInit {
			lockerInfo.LockedNeoHash = request.GetNep5TxHash()
		} else {
			if lockerInfo.Fail {
				return nil, fmt.Errorf("lock fail: %s", lockerInfo.Remark)
			}
		}
	} else {
		// init info
		info := &types.LockerInfo{
			State:         types.DepositInit,
			RHash:         request.GetRHash(),
			LockedNeoHash: request.GetNep5TxHash(),
		}
		if err := d.store.AddLockerInfo(info); err != nil {
			d.logger.Error(err)
			return nil, err
		}
		d.logger.Infof("add [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositInit))
	}

	go func() {
		lock(request.GetRHash(), d.logger)
		defer unlock(request.GetRHash(), d.logger)

		var info *types.LockerInfo
		var err error
		defer func() {
			d.store.SetLockerStateFail(info, err)
		}()

		info, err = d.store.GetLockerInfo(request.GetRHash())
		if err != nil {
			d.logger.Error(err)
			return
		}
		if info.State >= types.DepositEthLockedPending {
			d.logger.Infof("[%s] state already ahead [%s]", request.GetRHash(), types.LockerStateToString(types.DepositEthLockedPending))
			return
		}

		height, err := d.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), d.cfg.NEOCfg.ConfirmedHeight, neo.UserLock)
		if err != nil {
			d.logger.Error(err)
			return
		}

		swapInfo, err := d.neo.QuerySwapInfo(request.GetRHash())
		if err != nil {
			d.logger.Errorf("query swap info: %s", err)
			return
		}
		d.logger.Infof("swap info: %s", util.ToString(swapInfo))

		if b, h := d.neo.HasConfirmedBlocksHeight(height, getLockDeadLineHeight(swapInfo.OvertimeBlocks)); b {
			err = fmt.Errorf("lock time deadline has been exceeded [%s] [%d -> %d]", info.RHash, height, h)
			d.logger.Error(err)
			return
		}

		info.State = types.DepositNeoLockedDone
		info.LockedNeoHeight = height
		info.Amount = swapInfo.Amount
		info.NeoUserAddr = swapInfo.UserNeoAddress
		info.NeoTimerInterval = swapInfo.OvertimeBlocks
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Error(err)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoLockedDone))

		// wrapper to eth lock
		tx, err := d.eth.WrapperLock(request.GetRHash(), d.cfg.EthereumCfg.SignerAddress, swapInfo.Amount)
		if err != nil {
			d.logger.Error(err)
			return
		}
		d.logger.Infof("deposit/wrapper eth lock: %s [%s]", request.GetRHash(), tx)
		info.State = types.DepositEthLockedPending
		info.LockedEthHash = tx
		info.EthTimerInterval = d.cfg.EthereumCfg.DepositInterval
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Error(err)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedPending))
	}()
	return toBoolean(true), nil
}

func (d *DepositAPI) checkLockParams(request *pb.DepositLockRequest) error {
	address := d.cfg.EthereumCfg.SignerAddress

	if address != request.GetAddr() {
		return fmt.Errorf("invalid wrapper eth address, want [%s], but get [%s]", address, request.GetAddr())
	}
	return nil
}

func (d *DepositAPI) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.Boolean, error) {
	d.logger.Info("api - deposit fetch: ", request.String())
	rHash := util.Sha256(request.GetROrigin())
	if info, err := d.store.GetLockerInfo(rHash); err != nil {
		//d.logger.Errorf("%s: %s", rHash, err)
		return nil, err
	} else if !info.NeoTimeout {
		//d.logger.Errorf("current state is %s, [%s], not timeout", types.LockerStateToString(info.State), info.RHash)
		return nil, fmt.Errorf("not yet timeout, state: %s", types.LockerStateToString(info.State))
	}
	go func() {
		var info *types.LockerInfo
		var err error
		defer func() {
			d.store.SetLockerStateFail(info, err)
		}()

		var swapInfo *neo.SwapInfo
		swapInfo, err = d.neo.QuerySwapInfo(rHash)
		if err != nil {
			d.logger.Errorf("query swap info: %s, [%s]", err, rHash)
			return
		}
		if swapInfo.UserNeoAddress != request.GetUserNep5Addr() {
			err = fmt.Errorf("invalid user nep5 address, %s, %s", swapInfo.UserNeoAddress, request.GetUserNep5Addr())
			d.logger.Error(err)
			return
		}

		var tx string
		tx, err = d.neo.RefundUser(request.ROrigin, d.cfg.NEOCfg.SignerAddress)
		if err != nil {
			d.logger.Errorf("refund user: %s [%s]", err, rHash)
			return
		}

		d.logger.Infof("deposit user fetch(neo): %s [%s] ", tx, rHash)

		info, _ = d.store.GetLockerInfo(rHash)
		info.State = types.DepositNeoFetchPending
		info.UnlockedNeoHash = tx
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Errorf("update locker info: %s [%s]", err, rHash)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoFetchPending))

		var height uint32
		height, err = d.neo.TxVerifyAndConfirmed(tx, d.cfg.NEOCfg.ConfirmedHeight)
		if err != nil {
			d.logger.Errorf("tx %s verify: %s [%s]", tx, err, rHash)
			return
		}
		info.UnlockedNeoHeight = height
		info.State = types.DepositNeoFetchDone
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Errorf("update locker info: %s [%s]", err, rHash)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoFetchDone))
	}()
	return toBoolean(true), nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}
