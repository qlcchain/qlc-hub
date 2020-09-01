package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"go.uber.org/zap"

	pb "github.com/qlcchain/qlc-hub/grpc/proto"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/store"
)

type DebugAPI struct {
	cfg          *config.Config
	ctx          context.Context
	nep5Account  *wallet.Account
	erc20Account common.Address
	store        *store.Store
	logger       *zap.SugaredLogger
}

func NewDebugAPI(ctx context.Context, cfg *config.Config) (*DebugAPI, error) {
	nep5Account, err := wallet.NewAccountFromWIF(cfg.NEOCfg.WIF)
	if err != nil {
		return nil, fmt.Errorf("NewDebugAPI/NewAccountFromWIF: %s", err)
	}
	_, address, err := eth.GetAccountByPriKey(cfg.EthereumCfg.Account)
	if err != nil {
		return nil, fmt.Errorf("NewDebugAPI/GetAccountByPriKey: %s", err)
	}
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	return &DebugAPI{
		ctx:          ctx,
		cfg:          cfg,
		nep5Account:  nep5Account,
		erc20Account: address,
		store:        store,
		logger:       log.NewLogger("api/debug"),
	}, nil
}

func (d DebugAPI) Debug(ctx context.Context, empty *empty.Empty) (*pb.String, error) {
	panic("implement me")
}
