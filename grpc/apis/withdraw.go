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
	bsc    *eth.Transaction
	store  *gorm.DB
	cfg    *config.Config
	ctx    context.Context
	logger *zap.SugaredLogger
}

func NewWithdrawAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, bsc *eth.Transaction, s *gorm.DB) *WithdrawAPI {
	api := &WithdrawAPI{
		cfg:    cfg,
		neo:    neo,
		store:  s,
		eth:    eth,
		bsc:    bsc,
		ctx:    ctx,
		logger: log.NewLogger("api/withdraw"),
	}
	go api.lister()
	go api.correctSwapPending()
	go api.correctSwapState()
	return api
}

func (w *WithdrawAPI) lister() {
	contractAddress := common.HexToAddress(w.cfg.EthCfg.EthNep5Contract)
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
			txHash := vLog.TxHash
			txHeight := vLog.BlockNumber
			w.logger.Infof("eth event: %s, %d", txHash.String(), txHeight)
			if event, err := filterer.ParseMint(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Hash := event.Nep5Hash
				neoHash := hex.EncodeToString(nep5Hash[:])

				if _, err := db.GetSwapInfoByTxHash(w.store, neoHash, types.NEO); err == nil {
					w.logger.Infof("deposit event, hash:%s, user:%s, amount:%s. neo[%s]",
						txHash.String(), user.String(), amount.String(), neoHash)
					go func() {
						if err := toConfirmDepositEthTx(txHash, txHeight, neoHash, user.String(), amount.Int64(),
							w.eth, w.cfg.EthCfg.EthConfirmedHeight, w.store, w.logger, true); err != nil {
							w.logger.Errorf("withdraw event: %s, eth tx[%s]", err, txHash.String())
						}
					}()
				}
			}
			if event, err := filterer.ParseBurn(vLog); event != nil && err == nil {
				user := event.User
				amount := event.Amount
				nep5Addr := event.Nep5Addr
				neoClient := w.neo.Client()
				if err := neoClient.ValidateAddress(nep5Addr); err == nil {
					w.logger.Infof("withdraw event, user:%s, amount:%s, nep5Addr:%s. eth[%s]",
						user.String(), amount.String(), nep5Addr, txHash.String())

					go func() {
						if err := w.toWaitConfirmWithdrawEthTx(txHash, txHeight, user, amount, nep5Addr, true, types.ETH); err != nil {
							w.logger.Errorf("withdraw event: %s, eth[%s]", err, txHash.String())
						}
					}()
				} else {
					w.logger.Errorf("withdraw event, invalid nep5 address: %s, %s, eth tx[%s]", err, nep5Addr, txHash.String())
				}
			}
		}
	}
}

func (w *WithdrawAPI) toWaitConfirmWithdrawEthTx(chainTxHash common.Hash, txHeight uint64, user common.Address, amount *big.Int, nep5Addr string, isEvent bool, chainType types.ChainType) error {
	if txHeight != 0 {
		if chainType == types.ETH {
			if err := w.eth.WaitTxVerifyAndConfirmed(chainTxHash, txHeight, w.cfg.EthCfg.EthConfirmedHeight+1); err != nil {
				return fmt.Errorf("eth tx confirmed: %s", err)
			}
		} else {
			if err := w.bsc.WaitTxVerifyAndConfirmed(chainTxHash, txHeight, w.cfg.BscCfg.BscConfirmedHeight+1); err != nil {
				return fmt.Errorf("bsc tx confirmed: %s", err)
			}
		}
	}

	lock(util.AddHashPrefix(chainTxHash.String()), w.logger)
	defer unlock(util.AddHashPrefix(chainTxHash.String()), w.logger)
	w.logger.Infof("withdraw eth transaction confirmed, %s, %t", chainTxHash.String(), isEvent)

	if _, err := db.GetSwapInfoByTxHash(w.store, chainTxHash.String(), types.ETH); err == nil {
		//w.logger.Errorf("confirmed eth tx repeatedly, %s", ethTxHash.String())
		return nil
	}

	swapInfo := &types.SwapInfo{
		State:       types.WithDrawPending,
		Amount:      amount.Int64(),
		EthTxHash:   chainTxHash.String(),
		NeoTxHash:   "",
		EthUserAddr: user.String(),
		NeoUserAddr: nep5Addr,
		Chain:       chainType,
		StartTime:   time.Now().Unix(),
	}
	w.logger.Infof("add state to %s, eth[%s]", types.SwapStateToString(types.WithDrawPending), chainTxHash.String())
	if err := db.InsertSwapInfo(w.store, swapInfo); err != nil {
		return fmt.Errorf("withdraw insert: %s", err)
	}

	neoTx, err := w.neo.CreateUnLockTransaction(chainTxHash.String(), nep5Addr, user.String(), int(amount.Int64()), w.cfg.NEOCfg.Owner)
	if err != nil {
		swapInfo.State = types.WithDrawFail
		db.UpdateSwapInfo(w.store, swapInfo)
		w.logger.Errorf("create neo tx: %s, neo[%s]", err, chainTxHash.String())
		return fmt.Errorf("create tx: %s", err)
	}

	w.logger.Infof("neo tx created: %s. eth[%s]", neoTx, chainTxHash.String())
	if _, err := w.neo.WaitTxVerifyAndConfirmed(neoTx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("neo tx confirmed: %s", err)
	}
	if _, err := w.neo.QueryLockedInfo(chainTxHash.String()); err != nil {
		return fmt.Errorf("cannot get swap info: %s", err)
	}
	w.logger.Infof("neo tx confirmed: %s, eth[%s]", neoTx, chainTxHash.String())
	swapInfo.NeoTxHash = neoTx
	swapInfo.State = types.WithDrawDone
	w.logger.Infof("update state to %s, eth[%s]", types.SwapStateToString(types.WithDrawDone), chainTxHash.String())
	if err := db.UpdateSwapInfo(w.store, swapInfo); err != nil {
		return err
	}
	w.logger.Infof("withdraw successfully, eth[%s]", chainTxHash.String())
	return nil
}

func toConfirmDepositEthTx(txHash common.Hash, txHeight uint64, neoTxHash string, ethUserAddr string, amount int64,
	chain *eth.Transaction, confirmedHeight int64, store *gorm.DB, logger *zap.SugaredLogger, isEvent bool) error {

	if txHeight != 0 {
		if err := chain.WaitTxVerifyAndConfirmed(txHash, txHeight, confirmedHeight+1); err != nil {
			return fmt.Errorf("tx confirmed: %s", err)
		}
	}

	lock(util.AddHashPrefix(txHash.String()), logger)
	defer unlock(util.AddHashPrefix(txHash.String()), logger)
	logger.Infof("deposit eth tx confirmed, %s, neo[%s], %t", txHash.String(), neoTxHash, isEvent)

	swapInfo, err := db.GetSwapInfoByTxHash(store, neoTxHash, types.NEO)
	if err != nil {
		logger.Error(err)
		return fmt.Errorf("get swapInfo: %s", err)
	}

	if swapInfo.State == types.DepositDone && swapInfo.EthTxHash != "" {
		return nil
	}

	swapInfo.State = types.DepositDone
	swapInfo.EthTxHash = txHash.String()
	swapInfo.EthUserAddr = ethUserAddr
	swapInfo.Amount = amount
	if err := db.UpdateSwapInfo(store, swapInfo); err != nil {
		logger.Error(err)
		return fmt.Errorf("set swapInfo: %s", err)
	}
	logger.Infof("update state to %s, neo[%s]", types.SwapStateToString(types.DepositDone), neoTxHash)
	logger.Infof("deposit successfully. neo[%s]", neoTxHash)
	return nil
}

func (w *WithdrawAPI) ChainTransactionConfirmed(ctx context.Context, request *pb.ChainTxRequest) (*pb.Boolean, error) {
	w.logger.Infof("call withdraw EthTransactionConfirmed: %s", request.String())
	hash := request.GetHash()
	if hash == "" {
		return nil, fmt.Errorf("invalid hash, %s", request)
	}
	chainType := types.StringToChainType(request.GetChainType())
	if chainType != types.ETH && chainType != types.BSC {
		return nil, errors.New("invalid chain")
	}

	swapInfo, err := db.GetSwapInfoByTxHash(w.store, hash, types.ETH)
	if err == nil && swapInfo != nil {
		if swapInfo.State == types.WithDrawDone {
			w.logger.Errorf("withdraw repeatedly, eth[%s]", hash)
			return nil, fmt.Errorf("withdraw repeatedly, tx[%s]", hash)
		}
		if swapInfo.State == types.WithDrawFail { // neo tx send fail
			go func() {
				neoTx, err := w.neo.CreateUnLockTransaction(swapInfo.EthTxHash, swapInfo.NeoUserAddr, swapInfo.EthUserAddr, int(swapInfo.Amount), w.cfg.NEOCfg.Owner)
				if err != nil {
					w.logger.Errorf("create neo unlock tx: %s", err)
					return
				}

				w.logger.Infof("neo tx created, %s. eth[%s]", neoTx, swapInfo.EthTxHash)
				if _, err := w.neo.WaitTxVerifyAndConfirmed(neoTx, w.cfg.NEOCfg.ConfirmedHeight); err != nil {
					w.logger.Errorf("neo unlock tx confirmed: %s", err)
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
		if swapInfo.State == types.WithDrawPending {
			w.logger.Infof("withdraw state is pending, eth[%s]", hash)
			return toBoolean(true), nil
		} else {
			return nil, errors.New("invalid state")
		}
	} else {
		var chain *eth.Transaction
		if chainType == types.ETH {
			chain = w.eth
		} else {
			chain = w.bsc
		}

		confirmed, err := chain.HasBlockConfirmed(common.HexToHash(hash), w.cfg.EthCfg.EthConfirmedHeight)
		if err != nil || !confirmed {
			w.logger.Infof("block not confirmed: %s, %s", err, hash)
			return nil, err
		}

		amount, user, nep5Addr, err := chain.SyncBurnLog(hash)
		if err != nil {
			w.logger.Error(err)
			return nil, err
		}
		if err := w.neo.ValidateAddress(nep5Addr); err != nil {
			w.logger.Error(err)
			return nil, err
		}
		w.logger.Infof("got burn log: user:%s, neoAddr:%s, amount:%d. [%s]", user.String(), nep5Addr, amount.Int64(), hash)
		go func() {
			if err := w.toWaitConfirmWithdrawEthTx(common.HexToHash(hash), 0, user, amount, nep5Addr, false, chainType); err != nil {
				w.logger.Error(err)
				return
			}
		}()
		return toBoolean(true), nil
	}
}

func (w *WithdrawAPI) ChainTransactionSent(ctx context.Context, request *pb.ChainTxRequest) (*pb.Boolean, error) {
	w.logger.Infof("call withdraw EthTransactionSent: %s", request.String())
	hash := request.GetHash()
	if hash == "" {
		return nil, fmt.Errorf("invalid hash, %s", request)
	}
	chainType := types.StringToChainType(request.GetChainType())
	if chainType != types.ETH && chainType != types.BSC {
		return nil, errors.New("invalid chain")
	}

	if _, err := db.GetSwapPendingByTxEthHash(w.store, hash); err != nil {
		if err := db.InsertSwapPending(w.store, &types.SwapPending{
			Typ:       types.Withdraw,
			EthTxHash: hash,
			Chain:     chainType,
		}); err != nil {
			w.logger.Error(err)
			return toBoolean(false), err
		}
	}

	go func() {
		var chain *eth.Transaction
		if chainType == types.ETH {
			chain = w.eth
		} else {
			chain = w.bsc
		}

		if err := chain.WaitTxVerifyAndConfirmed(common.HexToHash(hash), 0, w.cfg.EthCfg.EthConfirmedHeight); err != nil {
			w.logger.Errorf("tx confirmed: %s", err)
		}
		amount, user, nep5Addr, err := chain.SyncBurnLog(hash)
		if err != nil {
			w.logger.Error(err)
			return
		}
		if err := w.neo.ValidateAddress(nep5Addr); err != nil {
			w.logger.Error(err)
			return
		}
		if err := w.toWaitConfirmWithdrawEthTx(common.HexToHash(hash), 0, user, amount, nep5Addr, false, chainType); err != nil {
			w.logger.Error(err)
			return
		}
	}()
	return toBoolean(true), nil
}

var (
	maxRHashSize = 10240
	timeout      = 48 * time.Hour
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

func (w *WithdrawAPI) correctSwapPending() error {
	vTicker := time.NewTicker(6 * time.Minute)
	for {
		select {
		case <-vTicker.C:
			infos, err := db.GetSwapPendings(w.store, 0, 0)
			if err != nil {
				w.logger.Error(err)
				continue
			}
			for _, info := range infos {
				if info.Typ == types.Withdraw && time.Now().Unix()-info.LastModifyTime > 60*10 {
					swapInfo, err := db.GetSwapInfoByTxHash(w.store, info.EthTxHash, types.ETH)
					if err == nil {
						if swapInfo.State == types.WithDrawDone {
							_ = db.DeleteSwapPending(w.store, info)
						}
					} else {
						w.logger.Infof("continue withdraw, eth[%s]", info.EthTxHash)
						if _, err := w.ChainTransactionSent(context.Background(), &pb.ChainTxRequest{
							Hash:      info.EthTxHash,
							ChainType: types.ChainTypeToString(info.Chain),
						}); err != nil {
							w.logger.Error(err)
						}
					}
				}
			}
		}
	}
}

// update by state, withdrawPending or DepositPending
func (w *WithdrawAPI) correctSwapState() error {
	vTicker := time.NewTicker(8 * time.Minute)
	for {
		select {
		case <-vTicker.C:
			infos, err := db.GetSwapInfos(w.store, "", 0, 0)
			if err != nil {
				w.logger.Error(err)
				continue
			}
			for _, info := range infos {
				if info.State == types.WithDrawPending && time.Now().Unix()-info.LastModifyTime > 60*10 {
					lockedInfo, err := w.neo.QueryLockedInfo(info.EthTxHash)
					if err == nil && lockedInfo.Amount == info.Amount {
						info.State = types.WithDrawDone
						info.NeoTxHash = lockedInfo.Txid
						if err := db.UpdateSwapInfo(w.store, info); err == nil {
							w.logger.Infof("correct withdraw swap state: eth[%s]", info.EthTxHash)
						}
					}
				}
				if info.State == types.DepositPending && time.Now().Unix()-info.LastModifyTime > 60*10 {
					var amount *big.Int
					var err error
					if info.Chain == types.ETH {
						amount, err = w.eth.GetLockedAmountByNeoTxHash(info.NeoTxHash)
					} else {
						amount, err = w.bsc.GetLockedAmountByNeoTxHash(info.NeoTxHash)
					}
					if err == nil && amount.Int64() == info.Amount {
						info.State = types.DepositDone //can not get tx hash in eth contract
						if err := db.UpdateSwapInfo(w.store, info); err == nil {
							w.logger.Infof("correct deposit swap state: neo[%s]", info.NeoTxHash)
						}
					}
				}
				if info.State == types.WithDrawFail && time.Now().Unix()-info.LastModifyTime > 60*10 {
					if _, err := w.ChainTransactionConfirmed(context.Background(), &pb.ChainTxRequest{
						Hash:      info.EthTxHash,
						ChainType: types.ChainTypeToString(info.Chain),
					}); err != nil {
						w.logger.Error(err)
					}
				}
			}
		}
	}
}
