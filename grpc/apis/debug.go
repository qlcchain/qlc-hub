package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

type DebugAPI struct {
	logger *zap.SugaredLogger
}

func NewDebugAPI() *DebugAPI {
	return &DebugAPI{
		logger: log.NewLogger("api/debug"),
	}
}

func (d *DebugAPI) HashState(ctx context.Context, s *pb.String) (*pb.StateInfo, error) {
	panic("implement me")
}

func (d *DebugAPI) Ping(ctx context.Context, empty *empty.Empty) (*pb.PingResponse, error) {
	panic("implement me")
}
