package apis

import (
	"context"

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
	neo    *neo.Transaction
	eth    *ethclient.Client
	store  *store.Store
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *ethclient.Client) (*WithdrawAPI, error) {
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	return &WithdrawAPI{
		cfg:    cfg,
		neo:    neo,
		store:  store,
		eth:    eth,
		ctx:    ctx,
		logger: log.NewLogger("api/withdraw"),
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
		b, height, err := w.neo.TxVerifyAndConfirmed(request.GetNep5TxHash(), neoConfirmedHeight)
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

		tx, err := eth.WrapperUnlock(rHash, request.GetROrigin(), w.cfg.EthereumCfg.Account, w.cfg.EthereumCfg.Contract, w.eth)
		if err != nil {
			w.logger.Errorf("eth wrapper unlock: %s [%s]", err, rHash)
			return
		}
		w.logger.Info("[%s] withdraw wrapper eth unlock: ", rHash, tx)
		info.State = types.WithDrawEthUnlockPending
		info.UnlockedErc20Hash = tx
		if err := w.store.UpdateLockerInfo(info); err != nil {
			return
		}
		w.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
	}()

	return toBoolean(true), nil
}
