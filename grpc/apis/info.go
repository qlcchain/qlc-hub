package apis

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
)

type InfoAPI struct {
	cfg    *config.Config
	ctx    context.Context
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewInfoAPI(ctx context.Context, cfg *config.Config) (*InfoAPI, error) {
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	return &InfoAPI{
		ctx:    ctx,
		cfg:    cfg,
		store:  store,
		logger: log.NewLogger("api/info"),
	}, nil
}

func (i *InfoAPI) Ping(ctx context.Context, e *empty.Empty) (*pb.PingResponse, error) {
	nep5Account, err := wallet.NewAccountFromWIF(i.cfg.NEOCfg.WIF)
	if err != nil {
		return nil, err
	}
	_, address, err := eth.GetAccountByPriKey(i.cfg.EthereumCfg.Account)
	if err != nil {
		return nil, err
	}
	return &pb.PingResponse{
		NeoContract: i.cfg.NEOCfg.Contract,
		NeoAddress:  nep5Account.Address,
		EthContract: i.cfg.EthereumCfg.Contract,
		EthAddress:  address.String(),
	}, nil
}

func (i *InfoAPI) LockerInfo(ctx context.Context, s *pb.String) (*pb.LockerStateResponse, error) {
	r, err := i.store.GetLockerInfo(s.GetValue())
	if err != nil {
		return nil, err
	}
	return toLockerState(r), nil
}

func (i *InfoAPI) LockerInfosByErc20Addr(ctx context.Context, offset *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	if offset.GetCount() < 0 || offset.GetOffset() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetCount(), offset.GetOffset())
	}
	as := make([]*pb.LockerStateResponse, 0)
	err := i.store.GetLockerInfos(func(info *types.LockerInfo) error {
		if info.Erc20Addr == offset.GetParam() {
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
	states := getStateByOffset(as, offset.GetCount(), offset.GetOffset())
	return &pb.LockerStatesResponse{
		Lockers: states,
	}, nil
}

func (i *InfoAPI) LockerInfosByNep5Addr(ctx context.Context, offset *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	if offset.GetCount() < 0 || offset.GetOffset() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetCount(), offset.GetOffset())
	}
	as := make([]*pb.LockerStateResponse, 0)
	err := i.store.GetLockerInfos(func(info *types.LockerInfo) error {
		if info.Nep5Addr == offset.GetParam() {
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
	states := getStateByOffset(as, offset.GetCount(), offset.GetOffset())
	return &pb.LockerStatesResponse{
		Lockers: states,
	}, nil
}

func (i *InfoAPI) LockerInfos(ctx context.Context, offset *pb.Offset) (*pb.LockerStatesResponse, error) {
	if offset.GetCount() < 0 || offset.GetOffset() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetCount(), offset.GetOffset())
	}
	as := make([]*pb.LockerStateResponse, 0)
	err := i.store.GetLockerInfos(func(info *types.LockerInfo) error {
		as = append(as, toLockerState(info))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].LastModifyTime > as[j].LastModifyTime
	})
	states := getStateByOffset(as, offset.GetCount(), offset.GetOffset())
	return &pb.LockerStatesResponse{
		Lockers: states,
	}, nil
}

func getStateByOffset(states []*pb.LockerStateResponse, count, offset int32) []*pb.LockerStateResponse {
	length := int32(len(states))
	if length == 0 {
		return make([]*pb.LockerStateResponse, 0)
	}
	if count == 0 && offset == 0 {
		return states
	}
	if length < offset {
		return make([]*pb.LockerStateResponse, 0)
	}
	if length < offset+count {
		return states[offset:length]
	} else {
		return states[offset : offset+count]
	}
}

func toLockerState(s *types.LockerInfo) *pb.LockerStateResponse {
	return &pb.LockerStateResponse{
		State:               int64(s.State),
		StateStr:            types.LockerStateToString(s.State),
		RHash:               s.RHash,
		ROrigin:             s.ROrigin,
		Amount:              s.Amount,
		UserErc20Addr:       s.Erc20Addr,
		UserNep5Addr:        s.Nep5Addr,
		LockedNep5Hash:      s.LockedNep5Hash,
		LockedNep5Height:    s.LockedNep5Height,
		LockedErc20Hash:     s.LockedErc20Hash,
		LockedErc20Height:   s.LockedErc20Height,
		UnlockedNep5Hash:    s.UnlockedNep5Hash,
		UnlockedNep5Height:  s.UnlockedNep5Height,
		UnlockedErc20Hash:   s.UnlockedErc20Hash,
		UnlockedErc20Height: s.UnlockedErc20Height,
		StartTime:           time.Unix(s.StartTime, 0).Format(time.RFC3339),
		LastModifyTime:      time.Unix(s.LastModifyTime, 0).Format(time.RFC3339),
		Fail:                s.Fail,
		Remark:              s.Remark,
	}
}
