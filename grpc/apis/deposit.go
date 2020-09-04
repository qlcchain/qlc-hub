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
	"github.com/qlcchain/qlc-hub/pkg/util"
	"go.uber.org/zap"
)

type DepositAPI struct {
	eth    *ethclient.Client
	neo    *neo.Transaction
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewDepositAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client, s *store.Store) *DepositAPI {
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
			lockerInfo.LockedNep5Hash = request.GetNep5TxHash()
		} else {
			if lockerInfo.Fail {
				return nil, fmt.Errorf("lock fail: %s", lockerInfo.Remark)
			} else {
				return toBoolean(true), nil
			}
		}
	} else {
		// init info
		info := &types.LockerInfo{
			State:          types.DepositInit,
			RHash:          request.GetRHash(),
			LockedNep5Hash: request.GetNep5TxHash(),
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

		info, err := d.store.GetLockerInfo(request.GetRHash())
		if err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		if info.State >= types.DepositEthLockedPending {
			d.logger.Infof("[%s] state already ahead [%s]", request.GetRHash(), types.LockerStateToString(types.DepositEthLockedPending))
			return
		}

		height, err := d.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), d.cfg.NEOCfg.ConfirmedHeight, neo.UserLock)
		if err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}

		swapInfo, err := d.neo.QuerySwapInfo(request.GetRHash())
		if err != nil {
			d.logger.Errorf("query swap info: %s", err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		d.logger.Infof("swap info: %s", util.ToString(swapInfo))

		info.State = types.DepositNeoLockedDone
		info.LockedNep5Height = height
		info.Amount = swapInfo.Amount
		info.Nep5Addr = swapInfo.UserNeoAddress
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoLockedDone))

		// wrapper to eth lock
		tx, err := eth.WrapperLock(request.GetRHash(), d.cfg.EthereumCfg.Account, d.cfg.EthereumCfg.Contract, swapInfo.Amount, d.eth)
		if err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		d.logger.Infof("deposit/wrapper eth lock: %s [%s]", request.GetRHash(), tx)
		info.State = types.DepositEthLockedPending
		info.LockedErc20Hash = tx
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedPending))
	}()
	return toBoolean(true), nil
}

func (d *DepositAPI) checkLockParams(request *pb.DepositLockRequest) error {
	_, address, err := eth.GetAccountByPriKey(d.cfg.EthereumCfg.Account)
	if err != nil {
		return fmt.Errorf("invalid erc20 account: %s [%s]", err, d.cfg.EthereumCfg.Account)
	}
	if address.String() != request.GetAddr() {
		return fmt.Errorf("invalid wrapper eth address, want [%s], but get [%s]", address.String(), request.GetAddr())
	}
	return nil
}

func (d *DepositAPI) FetchNotice(ctx context.Context, request *pb.FetchNoticeRequest) (*pb.Boolean, error) {
	d.logger.Info("api - deposit fetchNotice: ", request.String())
	info, err := d.store.GetLockerInfo(request.GetRHash())
	if err != nil {
		d.logger.Errorf("%s: %s", request.GetRHash(), err)
		return nil, err
	}
	if !info.NeoTimeout {
		d.logger.Errorf("current [%s] is [%s], not timeout", info.RHash, types.LockerStateToString(info.State))
		return nil, fmt.Errorf("not yet timeout, state: %s", types.LockerStateToString(info.State))
	}
	go func() {
		height, err := d.neo.CheckTxAndRHash(request.GetNep5TxHash(), request.GetRHash(), d.cfg.NEOCfg.ConfirmedHeight, neo.RefundUser)
		if err != nil {
			d.logger.Error(err)
			d.store.SetLockerStateFail(info, err)
			return
		}

		info.State = types.DepositNeoFetchDone
		info.UnlockedNep5Height = height
		info.UnlockedNep5Hash = request.GetNep5TxHash()
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Errorf("%s: %s", request.GetRHash(), err)
			d.store.SetLockerStateFail(info, err)
			return
		}
		d.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoFetchDone))
	}()
	return toBoolean(true), nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}
