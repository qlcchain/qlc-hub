package apis

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
)

type DebugAPI struct {
	cfg    *config.Config
	eth    *eth.Transaction
	neo    *neo.Transaction
	ctx    context.Context
	store  *gorm.DB
	logger *zap.SugaredLogger
}

func NewDebugAPI(ctx context.Context, cfg *config.Config, eth *eth.Transaction, neo *neo.Transaction, s *gorm.DB) *DebugAPI {
	return &DebugAPI{
		ctx:    ctx,
		cfg:    cfg,
		eth:    eth,
		neo:    neo,
		store:  s,
		logger: log.NewLogger("api/debug"),
	}
}

func (d *DebugAPI) SignData(ctx context.Context, s *pb.String) (*pb.SignDataResponse, error) {
	r, err := d.neo.SignData(d.cfg.NEOCfg.OwnerAddress, s.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.SignDataResponse{
		Sign:       r.Sign,
		VerifyData: r.VerifyData,
	}, nil
}
