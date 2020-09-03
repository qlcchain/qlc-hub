package apis

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"go.uber.org/zap"
)

type DebugAPI struct {
	cfg    *config.Config
	eth    *ethclient.Client
	ctx    context.Context
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewDebugAPI(ctx context.Context, cfg *config.Config, eth *ethclient.Client, s *store.Store) *DebugAPI {
	return &DebugAPI{
		ctx:    ctx,
		cfg:    cfg,
		eth:    eth,
		store:  s,
		logger: log.NewLogger("api/debug"),
	}
}

func (d *DebugAPI) HashTimer(ctx context.Context, s *pb.String) (*pb.HashTimerResponse, error) {
	timer, err := eth.GetHashTimer(d.eth, d.cfg.EthereumCfg.Contract, s.GetValue())
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

func (d *DebugAPI) LockerInfosCount(ctx context.Context, e *empty.Empty) (*pb.LockerInfosCountResponse, error) {
	return nil, nil
}

func (d *DebugAPI) LockerInfosByState(ctx context.Context, s *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	return nil, nil
}
