package apis

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WithdrawAPI struct {
	neo    *neo.Transaction
	eth    *eth.Transaction
	store  *gorm.DB
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, s *gorm.DB) *WithdrawAPI {
	api := &WithdrawAPI{
		cfg:    cfg,
		neo:    neo,
		store:  s,
		eth:    eth,
		ctx:    ctx,
		logger: log.NewLogger("api/withdraw"),
	}
	go api.lister()
	return api
}

func (w *WithdrawAPI) lister() {
	contractAddress := common.HexToAddress(w.cfg.EthCfg.Contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	filterer, err := eth.NewQLCChainFilterer(contractAddress, w.eth.Client())
	if err != nil {
		w.logger.Error("NewQLCChainFilterer: ", err)
		return
	}
	logs := make(chan ethTypes.Log)
	sub, err := w.eth.Client().SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		w.logger.Error("SubscribeFilterLogs: ", err)
		return
	}
	for {
		select {
		case <-w.ctx.Done():
			return
		case err := <-sub.Err():
			if err != nil {
				w.logger.Error("SubscribeFilterLogs: ", err)
			}
		case vLog := <-logs:
			if event, err := filterer.ParseBurn(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Addr := event.Nep5Addr
				txHash := vLog.TxHash
				txHeight := vLog.BlockNumber

				w.logger.Infof("withdraw event, user:%s, amount:%s, nep5Addr:%s. eth tx[%s,%d]",
					user.String(), amount.String(), nep5Addr, txHash.String(), txHeight)

				if err := w.toConfirmWithdrawEthTx(txHash, txHeight, user, amount, nep5Addr); err != nil {
					w.logger.Errorf("withdraw event: %s, tx[%s]", err, txHash.String())
					continue
				}
			}
			if event, err := filterer.ParseMint(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Hash := event.Nep5Hash
				txHash := vLog.TxHash
				txHeight := vLog.BlockNumber
				nHash := hex.EncodeToString(nep5Hash[:])

				w.logger.Infof("deposit event, user:%s, amount:%s, nep5Hash:%s. eth tx[%s,%d]",
					user.String(), amount.String(), nHash, txHash.String(), txHeight)

				swapInfo, err := db.GetSwapInfoByTxHash(w.store, nHash, types.NEO)
				if err != nil {
					w.logger.Error(err)
					continue
				}
				if swapInfo.EthUserAddr != user.String() || swapInfo.Amount != amount.Int64() {
					w.logger.Errorf("wrong info: %s", nHash)
					continue
				}
				if err := w.toConfirmDepositEthTx(txHash, txHeight, nHash, user.String()); err != nil {
					w.logger.Errorf("withdraw event: %s, tx[%s]", err, txHash.String())
					continue
				}
			}
		}
	}
}

func (w *WithdrawAPI) toConfirmWithdrawEthTx(txHash common.Hash, txHeight uint64, user common.Address, amount *big.Int, nep5Addr string) error {
	if err := w.eth.TxVerifyAndConfirmed(txHash, txHeight, w.cfg.EthCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("tx confirmed: %s", err)
	}
	w.logger.Infof("withdraw eth tx[%s] confirmed", txHash.String())

	swapInfo := &types.SwapInfo{
		State:       types.WithDrawPending,
		Amount:      amount.Int64(),
		EthTxHash:   txHash.String(),
		NeoTxHash:   "",
		EthUserAddr: user.String(),
		NeoUserAddr: "",
		StartTime:   time.Now().Unix(),
	}
	return db.InsertSwapInfo(w.store, swapInfo)
}

func (w *WithdrawAPI) toConfirmDepositEthTx(txHash common.Hash, txHeight uint64, neoTxHash string, ethUserAddr string) error {
	if err := w.eth.TxVerifyAndConfirmed(txHash, txHeight, w.cfg.EthCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("tx confirmed: %s", err)
	}
	w.logger.Infof("deposit eth tx[%s] confirmed", txHash.String())

	swapInfo, err := db.GetSwapInfoByTxHash(w.store, neoTxHash, types.NEO)
	if err != nil {
		return fmt.Errorf("get swapInfo: %s", err)
	}
	swapInfo.State = types.DepositDone
	swapInfo.EthTxHash = txHash.String()
	swapInfo.EthUserAddr = ethUserAddr
	if err := db.UpdateSwapInfo(w.store, swapInfo); err != nil {
		return fmt.Errorf("set swapInfo: %s", err)
	}
	return nil
}
