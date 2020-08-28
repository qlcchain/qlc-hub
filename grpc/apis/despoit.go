package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

type DepositAPI struct {
	logger *zap.SugaredLogger
}

func NewDepositAPI() *DepositAPI {
	return &DepositAPI{
		logger: log.NewLogger("api/deposit"),
	}
}

func (d *DepositAPI) DepositLock(ctx context.Context, request *pb.DepositLockRequest) (*pb.Boolean, error) {
	panic("implement me")
}

func (d *DepositAPI) DepositFetchNotice(ctx context.Context, request *pb.DepositFetchNoticeRequest) (*pb.Boolean, error) {
	panic("implement me")
}

func (d *DepositAPI) DepositEvent(empty *empty.Empty, server pb.DepositAPI_DepositEventServer) error {
	panic("implement me")
}
