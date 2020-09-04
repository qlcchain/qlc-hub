package apis

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
)

type EventAPI struct {
	eth    *ethclient.Client
	neo    *neo.Transaction
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewEventAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client, s *store.Store) *EventAPI {
	api := &EventAPI{
		cfg:    cfg,
		eth:    eth,
		neo:    neo,
		store:  s,
		ctx:    ctx,
		logger: log.NewLogger("api/event"),
	}
	go api.ethEventLister()
	go api.loopLockerState()
	return api
}

func (e *EventAPI) Event(empty *empty.Empty, server pb.EventAPI_EventServer) error {
	return nil
}
