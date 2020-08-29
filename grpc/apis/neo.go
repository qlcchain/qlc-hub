package apis

import (
	"context"
	"fmt"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"go.uber.org/zap"
)

type NeoAPI struct {
	neoTransaction *neo.NeoTransaction
	cfg            *config.Config
	ctx            context.Context
	logger         *zap.SugaredLogger
}

func NewNeoAPI(ctx context.Context, cfg *config.Config) (*NeoAPI, error) {
	nt, err := neo.NewNeoTransaction(cfg.NEOCfg.EndPoint, cfg.NEOCfg.Contract)
	if err != nil {
		return nil, fmt.Errorf("neo transaction: %s", err)
	}
	return &NeoAPI{
		cfg:            cfg,
		neoTransaction: nt,
		ctx:            ctx,
		logger:         log.NewLogger("api/deposit"),
	}, nil
}

func (d *NeoAPI) DepositFetchNotice(ctx context.Context, request *pb.FetchNoticeRequest) (*pb.Boolean, error) {
	panic("implement me")
}
