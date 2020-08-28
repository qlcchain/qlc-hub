package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

type WithdrawAPI struct {
	logger *zap.SugaredLogger
}

func NewWithdrawAPI() *WithdrawAPI {
	return &WithdrawAPI{
		logger: log.NewLogger("api/withdraw"),
	}
}

func (w WithdrawAPI) WithDrawUnlock(ctx context.Context, request *pb.WithDrawUnlockRequest) (*pb.Boolean, error) {
	panic("implement me")
}

func (w WithdrawAPI) WithDrawEvent(empty *empty.Empty, server pb.WithDrawAPI_WithDrawEventServer) error {
	panic("implement me")
}
