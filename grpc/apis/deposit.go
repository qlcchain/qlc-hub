package apis

import (
	"context"
	"errors"
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
			if err := d.store.UpdateLockerInfo(lockerInfo); err != nil {
				return nil, err
			}
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
		if info.State >= types.DepositEthLockedPending { // user can call repeatedly before DepositEthLockedPending
			d.logger.Infof("locker state already ahead of %s [%s] ", types.LockerStateToString(types.DepositEthLockedPending), request.GetRHash())
			return
		}

		_, err = d.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), d.cfg.NEOCfg.ConfirmedHeight, neo.UserLock)
		if err != nil {
			d.logger.Error(err)
			return
		}

		var swapInfo *neo.SwapInfo
		swapInfo, err = d.neo.QuerySwapInfoAndConfirmedTx(request.GetRHash(), neo.UserLock, d.cfg.NEOCfg.ConfirmedHeight)
		if err != nil {
			d.logger.Errorf("query swap info: %s [%s]", err, request.GetRHash())
			return
		}
		d.logger.Infof("swap info: %s", util.ToString(swapInfo))

		info.State = types.DepositNeoLockedDone
		info.LockedNeoHeight = swapInfo.LockedHeight
		info.LockedNeoHash = swapInfo.TxIdIn
		info.Amount = swapInfo.Amount
		info.NeoUserAddr = swapInfo.UserNeoAddress
		info.NeoTimerInterval = swapInfo.OvertimeBlocks
		if err := d.store.UpdateLockerInfo(info); err != nil {
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoLockedDone))

		//set state to DepositNeoLockedDone first, than if locker info is incorrect, user can call fetch
		if b, h := d.neo.HasConfirmedBlocksHeight(swapInfo.LockedHeight, getLockDeadLineHeight(swapInfo.OvertimeBlocks)); b {
			err = fmt.Errorf("lock time deadline has been exceeded [%s] [%d -> %d]", info.RHash, swapInfo.LockedHeight, h)
			d.logger.Error(err)
			return
		}

		if swapInfo.Amount < d.cfg.MinDepositAmount {
			err = fmt.Errorf("deposit locked amount %d should more than %d [%s]", swapInfo.Amount, d.cfg.MinDepositAmount, request.GetRHash())
			d.logger.Error(err)
			return
		}

		// wrapper to eth lock
		var tx string
		tx, err = d.eth.WrapperLock(request.GetRHash(), d.cfg.EthereumCfg.SignerAddress, swapInfo.Amount)
		if err != nil {
			d.logger.Error(err)
			return
		}
		d.logger.Infof("deposit/wrapper eth lock tx: %s [%s]", request.GetRHash(), tx)
		info.State = types.DepositEthLockedPending
		info.EthTimerInterval = d.cfg.EthereumCfg.DepositInterval
		if err := d.store.UpdateLockerInfo(info); err != nil {
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

func (d *DepositAPI) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.String, error) {
	rHash := util.Sha256(request.GetROrigin())
	d.logger.Infof("api - deposit fetch: %s, [%s]", request.String(), rHash)
	if err := d.neo.ValidateAddress(request.GetUserNep5Addr()); err != nil {
		return nil, fmt.Errorf("invalid address: %s", request.GetUserNep5Addr())
	}

	info, err := d.store.GetLockerInfo(rHash)
	if err != nil {
		return nil, err
	}
	if !info.NeoTimeout {
		return nil, fmt.Errorf("not yet timeout, state: %s", types.LockerStateToString(info.State))
	}
	if info.State == types.DepositNeoFetchDone {
		return toString(info.UnlockedNeoHash), nil
	}

	swapInfo, err := d.neo.QuerySwapInfo(rHash)
	if err != nil {
		d.logger.Errorf("query swap info: %s, [%s]", err, rHash)
		return nil, err
	}
	if swapInfo.UserNeoAddress != request.GetUserNep5Addr() {
		d.logger.Errorf("invalid user nep5 address, %s, %s [%s]", swapInfo.UserNeoAddress, request.GetUserNep5Addr(), rHash)
		return nil, errors.New("nep5 addr not match")
	}

	tx, err := d.neo.RefundUser(request.ROrigin, d.cfg.NEOCfg.SignerAddress)
	if err != nil {
		d.logger.Errorf("refund user: %s [%s]", err, rHash)
		return nil, err
	}
	d.logger.Infof("deposit user fetch(neo): %s [%s] ", tx, rHash)

	info.State = types.DepositNeoFetchPending
	if err := d.store.UpdateLockerInfo(info); err != nil {
		return nil, err
	}
	d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoFetchPending))

	go func() {
		lock(rHash, d.logger)
		defer unlock(rHash, d.logger)

		if err := setDepositNeoFetchDone(info.RHash, d.neo, d.store, d.cfg.NEOCfg.ConfirmedHeight, true, d.logger); err != nil {
			d.logger.Error(err)
			return
		}
	}()
	return toString(tx), nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}

func toString(b string) *pb.String {
	return &pb.String{Value: b}
}
