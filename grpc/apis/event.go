package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
)

type EventAPI struct {
	ethContract string
	eth         *ethclient.Client
	neo         *neo.Transaction
	nep5Account *wallet.Account
	store       *store.Store
	cfg         *config.Config
	ctx         context.Context
	logger      *zap.SugaredLogger
}

func NewEventAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client) (*EventAPI, error) {
	nep5Account, err := wallet.NewAccountFromWIF(cfg.NEOCfg.WIF)
	if err != nil {
		return nil, fmt.Errorf("NewAccountFromWIF: %s", err)
	}
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	api := &EventAPI{
		cfg:         cfg,
		ethContract: cfg.EthereumCfg.Contract,
		eth:         eth,
		neo:         neo,
		nep5Account: nep5Account,
		store:       store,
		ctx:         ctx,
		logger:      log.NewLogger("api/event"),
	}
	go api.ethEventLister()
	go api.loopLockerState()
	return api, nil
}

func (e *EventAPI) Event(empty *empty.Empty, server pb.EventAPI_EventServer) error {
	panic("implement me")
}
