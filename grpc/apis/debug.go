package apis

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

type DebugAPI struct {
	cfg    *config.Config
	eth    *eth.Transaction
	neo    *neo.Transaction
	ctx    context.Context
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewDebugAPI(ctx context.Context, cfg *config.Config, eth *eth.Transaction, neo *neo.Transaction, s *store.Store) *DebugAPI {
	return &DebugAPI{
		ctx:    ctx,
		cfg:    cfg,
		eth:    eth,
		neo:    neo,
		store:  s,
		logger: log.NewLogger("api/debug"),
	}
}

func (d *DebugAPI) HashTimer(ctx context.Context, s *pb.String) (*pb.HashTimerResponse, error) {
	timer, err := d.eth.GetHashTimer(s.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.HashTimerResponse{
		RHash:          timer.RHash,
		ROrigin:        timer.ROrigin,
		Amount:         timer.Amount.Int64(),
		UserAddr:       timer.UserAddr,
		LockedHeight:   timer.LockedHeight,
		UnlockedHeight: timer.UnlockedHeight,
	}, nil
}

func (d *DebugAPI) LockerInfosCount(ctx context.Context, e *empty.Empty) (*pb.LockerInfosStatResponse, error) {
	stat := make(map[string]*pb.LockerInfosStat)
	stat["Total"] = new(pb.LockerInfosStat)
	if err := d.store.GetLockerInfos(func(info *types.LockerInfo) error {
		stateKey := types.LockerStateToString(info.State)
		if _, ok := stat[stateKey]; !ok {
			stat[stateKey] = new(pb.LockerInfosStat)
		}
		stat[stateKey].Count = stat[stateKey].Count + 1
		stat[stateKey].Amount = stat[stateKey].Amount + info.Amount
		stat["Total"].Count = stat["Total"].Count + 1
		stat["Total"].Amount = stat["Total"].Amount + info.Amount
		return nil
	}); err != nil {
		return nil, err
	}
	return &pb.LockerInfosStatResponse{
		Result: stat,
	}, nil
}

func (d *DebugAPI) LockerInfosByState(ctx context.Context, params *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	if params.GetCount() < 0 || params.GetOffset() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", params.GetCount(), params.GetOffset())
	}
	as := make([]*pb.LockerStateResponse, 0)
	err := d.store.GetLockerInfos(func(info *types.LockerInfo) error {
		if types.LockerStateToString(info.State) == params.GetValue() {
			as = append(as, toLockerState(info))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].LastModifyTime > as[j].LastModifyTime
	})
	states := getStateByOffset(as, params.GetCount(), params.GetOffset())

	return &pb.LockerStatesResponse{
		Lockers: states,
	}, nil
}

func (d *DebugAPI) InterruptLocker(ctx context.Context, s *pb.LockerInterrupt) (*pb.Boolean, error) {
	locker, err := d.store.GetLockerInfo(s.GetRHash())
	if err != nil {
		return nil, err
	}
	locker.Interruption = s.GetInterrupt()
	locker.State = types.LockerState(s.GetState())
	if err := d.store.UpdateLockerInfo(locker); err != nil {
		return nil, err
	}
	return toBoolean(true), nil
}

func (d *DebugAPI) DeleteLockerInfo(ctx context.Context, request *pb.DeleteLockerInfoRequest) (*pb.Boolean, error) {
	d.logger.Infof("api -> delete : %s", request.String())
	var confirmedHeight uint32 = 10
	var gasTimes int32 = 2
	var maxCount int32 = 20

	if request.GetConfirmedHeight() > 0 {
		confirmedHeight = request.GetConfirmedHeight()
	}
	if request.GetMaxCount() < maxCount {
		maxCount = request.GetMaxCount()
	}
	if request.GetGasTimes() > 0 {
		gasTimes = request.GetGasTimes()
	}

	rHashes := make([]string, 0)
	if request.GetRHash() != "" {
		_, err := d.store.GetLockerInfo(request.GetRHash())
		if err != nil {
			return nil, err
		}
		rHashes = append(rHashes, request.GetRHash())
		d.deleteLockerInfos(rHashes)
	} else {
		currentGas, err := d.eth.GetBestGas()
		if err != nil {
			return nil, err
		}
		currentHeight, err := d.eth.GetBestBlockHeight()
		if err != nil {
			return nil, err
		}
		if err := d.store.GetLockerInfos(func(info *types.LockerInfo) error {
			if info.State == types.DepositNeoUnLockedDone || info.State == types.DepositNeoFetchDone ||
				info.State == types.WithDrawEthUnlockDone || info.State == types.WithDrawEthFetchDone {
				if (uint32(currentHeight)-info.UnlockedEthHeight > confirmedHeight) &&
					(currentGas.Int64()/info.GasPrice >= int64(gasTimes)) &&
					info.Deleted == types.NotDeleted {
					if len(rHashes) <= int(maxCount) {
						rHashes = append(rHashes, info.RHash)
					}
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
		go d.deleteLockerInfos(rHashes)
	}
	return toBoolean(true), nil
}

func (d *DebugAPI) deleteLockerInfos(rHashes []string) {
	for _, rHash := range rHashes {
		info, err := d.store.GetLockerInfo(rHash)
		if err != nil {
			return
		}
		if info.Deleted == types.DeletedDone {
			continue
		}
		d.logger.Warnf("delete locker info [%s]", rHash)
		info.Deleted = types.DeletedPending
		info.DeletedTime = time.Now().Unix()
		if err := d.store.UpdateLockerInfo(info); err != nil {
			d.logger.Errorf("delete locker: %s [%s]", err, rHash)
			continue
		}
		tx, err := d.neo.DeleteSwapInfo(info.RHash, d.cfg.NEOCfg.SignerAddress)
		if err != nil {
			d.logger.Errorf("delete locker: %s [%s]", err, rHash)
			continue
		}
		d.logger.Warnf("neo swap info deleted, tx: %s  [%s]", tx, rHash)

		tx, err = d.eth.DeleteHashTimer(info.RHash, d.cfg.EthereumCfg.OwnerAddress)
		if err != nil {
			d.logger.Errorf("delete locker: %s [%s]", err, rHash)
			continue
		}
		d.logger.Warnf("eth hash timer deleted, tx: %s [%s]", tx, rHash)
	}
}

func (d *DebugAPI) LockerInfosByDeletedState(ctx context.Context, params *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	if params.GetCount() < 0 || params.GetOffset() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", params.GetCount(), params.GetOffset())
	}
	as := make([]*pb.LockerStateResponse, 0)
	err := d.store.GetLockerInfos(func(info *types.LockerInfo) error {
		if types.LockerDeletedToString(info.Deleted) == params.GetValue() {
			as = append(as, toLockerState(info))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].LastModifyTime > as[j].LastModifyTime
	})
	states := getStateByOffset(as, params.GetCount(), params.GetOffset())

	return &pb.LockerStatesResponse{
		Lockers: states,
	}, nil
}

func (d *DebugAPI) SignData(ctx context.Context, s *pb.String) (*pb.SignResponse, error) {
	return d.neo.SignData(d.cfg.NEOCfg.SignerAddress, s.GetValue())
}
