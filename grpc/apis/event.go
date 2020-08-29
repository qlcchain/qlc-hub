package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"go.uber.org/zap"
)

type EventAPI struct {
	contractAddr   string
	ethClient      *ethclient.Client
	neoTransaction *neo.NeoTransaction
	cfg            *config.Config
	ctx            context.Context
	logger         *zap.SugaredLogger
}

func NewEventAPI(ctx context.Context, cfg *config.Config) (*EventAPI, error) {
	ethClient, err := ethclient.Dial(cfg.EthereumCfg.EndPoint)
	if err != nil {
		return nil, fmt.Errorf("eth client dail: %s", err)
	}
	nt, err := neo.NewNeoTransaction(cfg.NEOCfg.EndPoint, cfg.NEOCfg.Contract)
	if err != nil {
		return nil, fmt.Errorf("neo transaction: %s", err)
	}
	api := &EventAPI{
		cfg:            cfg,
		contractAddr:   cfg.EthereumCfg.Contract,
		ethClient:      ethClient,
		neoTransaction: nt,
		ctx:            ctx,
		logger:         log.NewLogger("api/event"),
	}
	api.ethEventLister()
	return api, nil
}

func (e *EventAPI) Event(empty *empty.Empty, server pb.EventAPI_EventServer) error {
	panic("implement me")
}
