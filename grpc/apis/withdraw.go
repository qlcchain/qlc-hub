package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
)

type WithdrawAPI struct {
	neoTransaction *neo.Transaction
	ethClient      *ethclient.Client
	store          *store.Store
	cfg            *config.Config
	ctx            context.Context
	logger         *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config) (*WithdrawAPI, error) {
	nt, err := neo.NewTransaction(cfg.NEOCfg.EndPoint, cfg.NEOCfg.Contract)
	if err != nil {
		return nil, fmt.Errorf("neo transaction: %s", err)
	}
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	ethClient, err := ethclient.Dial(cfg.EthereumCfg.EndPoint)
	if err != nil {
		return nil, fmt.Errorf("eth client dail: %s", err)
	}
	return &WithdrawAPI{
		cfg:            cfg,
		neoTransaction: nt,
		store:          store,
		ethClient:      ethClient,
		ctx:            ctx,
		logger:         log.NewLogger("api/withdraw"),
	}, nil
}

func (w *WithdrawAPI) Unlock(ctx context.Context, request *pb.WithdrawUnlockRequest) (*pb.Boolean, error) {
	w.logger.Info("api - withdraw unlock: ", request.String())
	//todo check params

	rHash := request.GetRHash()
	info, err := w.store.GetLockerInfo(rHash)
	if err != nil {
		return nil, err
	}
	go func() {
		b, height, err := neo.TxVerifyAndConfirmed(request.GetNep5TxHash(), neoConfirmedHeight, w.neoTransaction)
		if !b || err != nil {
			w.logger.Errorf("neo tx confirmed: %s, %v [%s]", err, b, rHash)
			return
		}
		info.State = types.WithDrawNeoUnLockedDone
		info.UnlockedNep5Height = height
		info.UnlockedNep5Hash = request.GetNep5TxHash()
		info.ROrigin = request.GetROrigin()
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoUnLockedDone))

		tx, err := eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract, w.ethClient)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			return
		}
		w.logger.Info("withdraw wrapper eth unlock: ", tx)
		info.State = types.WithDrawEthUnlockPending
		info.UnlockedErc20Hash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
	}()

	return toBoolean(true), nil
}
