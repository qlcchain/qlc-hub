package apis

import (
	"context"
	"fmt"
	"sort"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
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

func (d *DebugAPI) DeleteLockerInfo(ctx context.Context, s *pb.String) (*pb.Boolean, error) {
	panic("implement me")
}

func (d *DebugAPI) SignData(ctx context.Context, s *pb.String) (*pb.SignResponse, error) {
	return d.neo.SignData(d.cfg.NEOCfg.SignerAddress, s.GetValue())
}
