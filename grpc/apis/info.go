package apis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

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

type InfoAPI struct {
	eth    *eth.Transaction
	neo    *neo.Transaction
	cfg    *config.Config
	ctx    context.Context
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewInfoAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, s *store.Store) *InfoAPI {
	return &InfoAPI{
		ctx:    ctx,
		cfg:    cfg,
		neo:    neo,
		eth:    eth,
		store:  s,
		logger: log.NewLogger("api/info"),
	}
}

func (i *InfoAPI) Ping(ctx context.Context, s *pb.String) (*pb.PingResponse, error) {
	ethBalance, err := i.eth.Balance(i.cfg.EthereumCfg.OwnerAddress)
	if err != nil {
		return nil, err
	}
	eb, err := strconv.ParseFloat(fmt.Sprintf("%.8f", float64(ethBalance)/1e18), 64)
	if err != nil {
		return nil, err
	}
	neoBalance, err := i.neo.Balance(i.cfg.NEOCfg.AssetsAddress, i.cfg.NEOCfg.AssetId)
	if err != nil {
		return nil, err
	}
	nb, err := strconv.ParseFloat(fmt.Sprintf("%.8f", float64(neoBalance)/1e8), 64)
	if err != nil {
		return nil, err
	}
	return &pb.PingResponse{
		EthContract:       i.cfg.EthereumCfg.Contract,
		EthAddress:        i.cfg.EthereumCfg.OwnerAddress,
		NeoContract:       i.cfg.NEOCfg.Contract,
		NeoAddress:        i.cfg.NEOCfg.SignerAddress,
		EthBalance:        float32(eb),
		NeoBalance:        float32(nb),
		WithdrawLimit:     isWithdrawLimitExceeded(s.GetValue()),
		MinDepositAmount:  i.cfg.MinDepositAmount,
		MinWithdrawAmount: i.cfg.MinDepositAmount,
	}, nil
}

func (i *InfoAPI) LockerInfo(ctx context.Context, s *pb.String) (*pb.LockerStateResponse, error) {
	rHash := util.RemoveHexPrefix(s.GetValue())
	r, err := i.store.GetLockerInfo(rHash)
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
		if info.EthUserAddr == offset.GetValue() {
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
		if info.NeoUserAddr == offset.GetValue() {
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
		State:             int64(s.State),
		StateStr:          types.LockerStateToString(s.State),
		RHash:             s.RHash,
		ROrigin:           s.ROrigin,
		Amount:            s.Amount,
		LockedNeoHash:     s.LockedNeoHash,
		LockedNeoHeight:   s.LockedNeoHeight,
		LockedEthHash:     s.LockedEthHash,
		LockedEthHeight:   s.LockedEthHeight,
		UnlockedNeoHash:   s.UnlockedNeoHash,
		UnlockedNeoHeight: s.UnlockedNeoHeight,
		UnlockedEthHash:   s.UnlockedEthHash,
		UnlockedEthHeight: s.UnlockedEthHeight,
		NeoTimerInterval:  uint32(s.NeoTimerInterval),
		EthTimerInterval:  uint32(s.EthTimerInterval),
		StartTime:         time.Unix(s.StartTime, 0).Format(time.RFC3339),
		LastModifyTime:    time.Unix(s.LastModifyTime, 0).Format(time.RFC3339),
		NeoUserAddr:       s.NeoUserAddr,
		EthUserAddr:       s.EthUserAddr,
		GasPrice:          s.GasPrice,
		NeoTimeout:        s.NeoTimeout,
		EthTimeout:        s.EthTimeout,
		Fail:              s.Fail,
		Remark:            s.Remark,
		Interruption:      s.Interruption,
		Deleted:           types.LockerDeletedToString(s.Deleted),
		DeletedTime:       time.Unix(s.DeletedTime, 0).Format(time.RFC3339),
	}
}
