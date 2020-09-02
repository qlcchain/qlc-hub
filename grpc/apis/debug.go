package apis

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	pb "github.com/qlcchain/qlc-hub/grpc/proto"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/store"
)

type DebugAPI struct {
	cfg    *config.Config
	ctx    context.Context
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewDebugAPI(ctx context.Context, cfg *config.Config, s *store.Store) *DebugAPI {
	return &DebugAPI{
		ctx:    ctx,
		cfg:    cfg,
		store:  s,
		logger: log.NewLogger("api/debug"),
	}
}

func (d *DebugAPI) LockerInfosCount(ctx context.Context, e *empty.Empty) (*pb.LockerInfosCountResponse, error) {
	return nil, nil
}

func (d *DebugAPI) LockerInfosByState(ctx context.Context, s *pb.ParamAndOffset) (*pb.LockerStatesResponse, error) {
	return nil, nil
}
