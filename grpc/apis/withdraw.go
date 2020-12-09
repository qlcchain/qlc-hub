package apis

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
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
			if event, err := filterer.ParseMint(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Hash := event.Nep5Hash
				txHash := vLog.TxHash
				txHeight := vLog.BlockNumber
				neoHash := hex.EncodeToString(nep5Hash[:])

				if _, err := db.GetSwapInfoByTxHash(w.store, neoHash, types.NEO); err == nil {
					w.logger.Infof("deposit event, hash:%s, user:%s, amount:%s. neo[%s]",
						txHash.String(), user.String(), amount.String(), neoHash)
					if err := w.toConfirmDepositEthTx(txHash, txHeight, neoHash, user.String(), amount.Int64()); err != nil {
						w.logger.Errorf("withdraw event: %s, eth tx[%s]", err, txHash.String())
						continue
					}
					w.logger.Infof("deposit successfully. neo[%s]", neoHash)
				}
			}
			if event, err := filterer.ParseBurn(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Addr := event.Nep5Addr
				txHash := vLog.TxHash
				txHeight := vLog.BlockNumber
				neoClient := w.neo.Client()
				if err := neoClient.ValidateAddress(nep5Addr); err == nil {
					w.logger.Infof("withdraw event, user:%s, amount:%s, nep5Addr:%s. eth[%s,%d]",
						user.String(), amount.String(), nep5Addr, txHash.String(), txHeight)

					if err := w.toWaitConfirmWithdrawEthTx(txHash, txHeight, user, amount, nep5Addr); err != nil {
						w.logger.Errorf("withdraw event: %s, eth[%s]", err, txHash.String())
						continue
					}
					w.logger.Infof("withdraw successfully, eth[%s]", txHash.String())
				} else {
					w.logger.Errorf("withdraw event, invalid nep5 address: %s, %s, eth tx[%s]", err, nep5Addr, txHash.String())
				}
			}
		}
	}
}

func (w *WithdrawAPI) toWaitConfirmWithdrawEthTx(ethTxHash common.Hash, txHeight uint64, user common.Address, amount *big.Int, nep5Addr string) error {
	lock(util.AddHashPrefix(ethTxHash.String()), w.logger)
	defer unlock(util.AddHashPrefix(ethTxHash.String()), w.logger)

	if _, err := db.GetSwapInfoByTxHash(w.store, ethTxHash.String(), types.ETH); err == nil {
		w.logger.Errorf("confirmed eth tx repeatedly, %s", ethTxHash.String())
		return errors.New("confirmed eth tx repeatedly")
	}
	if txHeight != 0 {
		if err := w.eth.WaitTxVerifyAndConfirmed(ethTxHash, txHeight, w.cfg.EthCfg.ConfirmedHeight); err != nil {
			return fmt.Errorf("tx confirmed: %s", err)
		}
	}
	w.logger.Infof("withdraw eth tx confirmed. eth[%s]", ethTxHash.String())

	swapInfo := &types.SwapInfo{
		State:       types.WithDrawPending,
		Amount:      amount.Int64(),
		EthTxHash:   ethTxHash.String(),
		NeoTxHash:   "",
		EthUserAddr: user.String(),
		NeoUserAddr: nep5Addr,
		StartTime:   time.Now().Unix(),
	}
	w.logger.Infof("add state to %s, eth[%s]", types.SwapStateToString(types.WithDrawPending), ethTxHash.String())
	if err := db.InsertSwapInfo(w.store, swapInfo); err != nil {
		return fmt.Errorf("withdraw insert: %s", err)
	}

	neoTx, err := w.neo.CreateUnLockTransaction(ethTxHash.String(), nep5Addr, user.String(), int(amount.Int64()), w.cfg.NEOCfg.OwnerAddress)
	if err != nil {
		swapInfo.State = types.WithDrawFail
		db.UpdateSwapInfo(w.store, swapInfo)
		w.logger.Errorf("create neo tx: %s, neo[%s]", err, ethTxHash.String())
		return fmt.Errorf("create tx: %s", err)
	}

	w.logger.Infof("neo tx created: %s. eth[%s]", neoTx, ethTxHash.String())
	if _, err := w.neo.WaitTxVerifyAndConfirmed(neoTx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("tx confirmed: %s", err)
	}
	if _, err := w.neo.QueryLockedInfo(ethTxHash.String()); err != nil {
		return fmt.Errorf("cannot get swap info: %s", err)
	}
	w.logger.Infof("neo tx confirmed: %s, eth[%s]", neoTx, ethTxHash.String())
	swapInfo.NeoTxHash = neoTx
	swapInfo.State = types.WithDrawDone
	w.logger.Infof("update state to %s, eth[%s]", types.SwapStateToString(types.WithDrawDone), ethTxHash.String())
	return db.UpdateSwapInfo(w.store, swapInfo)
}

func (w *WithdrawAPI) toConfirmDepositEthTx(txHash common.Hash, txHeight uint64, neoTxHash string, ethUserAddr string, amount int64) error {
	if err := w.eth.WaitTxVerifyAndConfirmed(txHash, txHeight, w.cfg.EthCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("tx confirmed: %s", err)
	}
	w.logger.Infof("deposit eth tx confirmed, %s, neo[%s]", txHash.String(), neoTxHash)

	swapInfo, err := db.GetSwapInfoByTxHash(w.store, neoTxHash, types.NEO)
	if err != nil {
		w.logger.Error(err)
		return fmt.Errorf("get swapInfo: %s", err)
	}
	swapInfo.State = types.DepositDone
	swapInfo.EthTxHash = txHash.String()
	swapInfo.EthUserAddr = ethUserAddr
	swapInfo.Amount = amount
	if err := db.UpdateSwapInfo(w.store, swapInfo); err != nil {
		w.logger.Error(err)
		return fmt.Errorf("set swapInfo: %s", err)
	}
	w.logger.Infof("update state to %s, neo[%s]", types.SwapStateToString(types.DepositDone), neoTxHash)
	return nil
}

func (w *WithdrawAPI) EthTransactionConfirmed(ctx context.Context, h *pb.Hash) (*pb.Boolean, error) {
	w.logger.Infof("call withdraw EthTransactionConfirmed: %s", h.String())
	hash := h.GetHash()
	if hash == "" {
		return nil, fmt.Errorf("invalid hash, %s", h)
	}

	swapInfo, err := db.GetSwapInfoByTxHash(w.store, hash, types.ETH)
	if err == nil && swapInfo != nil {
		if swapInfo.State == types.WithDrawDone {
			w.logger.Errorf("withdraw repeatedly, eth[%s]", hash)
			return nil, fmt.Errorf("withdraw repeatedly, tx[%s]", hash)
		}
		if swapInfo.State == types.WithDrawFail { // neo tx send fail
			go func() {
				neoTx, err := w.neo.CreateUnLockTransaction(swapInfo.EthTxHash, swapInfo.NeoUserAddr, swapInfo.EthUserAddr, int(swapInfo.Amount), w.cfg.NEOCfg.OwnerAddress)
				if err != nil {
					w.logger.Errorf("create tx: %s", err)
					return
				}

				w.logger.Infof("neo tx created, %s. eth[%s]", neoTx, swapInfo.EthTxHash)
				if _, err := w.neo.WaitTxVerifyAndConfirmed(neoTx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
					w.logger.Errorf("tx confirmed: %s", err)
					return
				}
				if _, err := w.neo.QueryLockedInfo(swapInfo.EthTxHash); err != nil {
					w.logger.Errorf("cannot get swap info: %s", err)
					return
				}
				w.logger.Infof("neo tx confirmed: %s, eth[%s]", neoTx, swapInfo.EthTxHash)
				swapInfo.NeoTxHash = neoTx
				swapInfo.State = types.WithDrawDone
				w.logger.Infof("update state to %s, eth[%s]", types.SwapStateToString(types.WithDrawDone), swapInfo.EthTxHash)
				if err := db.UpdateSwapInfo(w.store, swapInfo); err != nil {
					w.logger.Error(err)
					return
				}
				w.logger.Infof("withdraw successfully, eth[%s]", hash)
			}()
			return toBoolean(true), nil
		}
		return toBoolean(true), nil
	} else {
		tx, p, err := w.eth.Client().TransactionByHash(context.Background(), common.HexToHash(hash))
		if tx != nil && !p && err == nil { // if tx not found , p is false
			amount, user, nep5Addr, err := w.eth.SyncBurnLog(hash)
			if err != nil {
				w.logger.Error(err)
				return nil, err
			}
			w.logger.Infof("got burn log: user:%s, neoAddr:%s, amount:%d. [%s]", user.String(), nep5Addr, amount.Int64(), hash)
			go func() {
				if err := w.toWaitConfirmWithdrawEthTx(common.HexToHash(hash), 0, user, amount, nep5Addr); err != nil {
					w.logger.Error(err)
					return
				}
				w.logger.Infof("withdraw successfully eth[%s]", hash)
			}()
			return toBoolean(true), nil
		} else {
			w.logger.Errorf("tx not confirmed, %s, %v,%v, eth[%s]", err, tx != nil, !p, hash)
			return toBoolean(false), fmt.Errorf("tx not confirmed")
		}
	}
}

var (
	maxRHashSize = 10240
	timeout      = 24 * time.Hour
)

var glock = gcache.New(maxRHashSize).Expiration(timeout).LRU().Build()

func lock(rHash string, logger *zap.SugaredLogger) {
	if v, err := glock.Get(rHash); err != nil {
		mutex := &sync.Mutex{}
		if err := glock.Set(rHash, mutex); err != nil {
			logger.Errorf("set lock fail: %s [%s]", err, rHash)
		}
		mutex.Lock()
	} else {
		if l, ok := v.(*sync.Mutex); ok {
			l.Lock()
		} else {
			logger.Errorf("invalid lock type [%s]", rHash)
		}
	}
}

func unlock(rHash string, logger *zap.SugaredLogger) {
	if v, err := glock.Get(rHash); err != nil {
		logger.Errorf("can not get lock: %s [%s]", err, rHash)
	} else {
		if l, ok := v.(*sync.Mutex); ok {
			l.Unlock()
		} else {
			logger.Errorf("invalid lock type [%s]", rHash)
		}
	}
}
