package apis

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

func (e *EventAPI) ethEventLister() {
	contractAddress := common.HexToAddress(e.cfg.EthereumCfg.Contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	filterer, err := eth.NewQLCChainFilterer(contractAddress, e.eth.Client())
	if err != nil {
		e.logger.Error("NewQLCChainFilterer: ", err)
		return
	}
	logs := make(chan ethTypes.Log)
	sub, err := e.eth.Client().SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		e.logger.Error("SubscribeFilterLogs: ", err)
		return
	}
	for {
		select {
		case <-e.ctx.Done():
			return
		case err := <-sub.Err():
			if err != nil {
				e.logger.Error("SubscribeFilterLogs: ", err)
			}
		case vLog := <-logs:
			event, err := filterer.ParseLockedState(vLog)
			if err != nil {
				continue
			}
			rHash := hex.EncodeToString(event.RHash[:])
			state := event.State.Int64()
			txHash := vLog.TxHash.Hex()
			txHeight := vLog.BlockNumber

			e.logger.Infof("[%d] event log: rHash[%s], txHash[%s], txHeight[%d]", state, rHash, txHash, txHeight)
			if eth.State(state) <= eth.DestroyFetch {
				go e.processEthEvent(state, rHash, txHash, txHeight)
			}
		}
	}
}

func (e *EventAPI) processEthEvent(state int64, rHash, tx string, txHeight uint64) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)

	var info *types.LockerInfo
	var err error
	defer func() {
		e.store.SetLockerStateFail(info, err)
	}()

	if info, err = e.store.GetLockerInfo(rHash); err != nil {
		e.logger.Errorf("event/get lockerInfo[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
		return
	}

	e.logger.Infof("event/waiting for eth event tx %s confirmed ", tx)
	if err = e.eth.TxVerifyAndConfirmed(tx, int64(txHeight), int64(e.cfg.EthereumCfg.ConfirmedHeight)); err != nil {
		e.logger.Errorf("event/txVerify(eth)[%d]: %s,  rHash[%s], txHash[%s]", state, err, rHash, tx)
		return
	}

	var hashTimer *eth.HashTimer
	if hashTimer, err = e.eth.GetHashTimer(rHash); err != nil {
		e.logger.Errorf("event/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
		return
	}

	switch eth.State(state) {
	case eth.IssueLock:
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.DepositEthLockedPending {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.DepositEthLockedPending), rHash)
			return
		}

		info.State = types.DepositEthLockedDone
		info.LockedEthHash = tx
		info.LockedEthHeight = hashTimer.LockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthLockedDone))

	case eth.IssueUnlock:
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.DepositEthLockedDone {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.DepositEthLockedDone), rHash)
			return
		}

		info.State = types.DepositEthUnLockedDone
		info.ROrigin = hashTimer.ROrigin
		info.UnlockedEthHash = tx
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		info.EthUserAddr = hashTimer.UserAddr
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

		// to neo unlock
		if err := setDepositNeoUnLockedPending(info.RHash, e.neo, e.store, e.cfg.NEOCfg.SignerAddress, e.logger); err != nil {
			e.logger.Errorf("set DepositNeoUnLockedPending: %s [%s -> %s]", err, hashTimer.ROrigin, info.RHash)
			return
		}

		if err := setDepositNeoUnLockedDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, true, e.logger); err != nil {
			e.logger.Errorf("set DepositNeoUnLockedDone: %s [%s]", err, info.RHash)
			return
		}

	case eth.IssueFetch: // wrapper Fetch
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.DepositEthFetchPending {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.DepositEthFetchPending), rHash)
			return
		}

		return
		info.State = types.DepositEthFetchDone
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		info.UnlockedEthHash = tx
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthFetchDone))

	case eth.DestroyLock:
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.WithDrawInit {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.WithDrawInit), rHash)
			return
		}

		info.State = types.WithDrawEthLockedDone
		info.RHash = rHash
		info.LockedEthHash = tx
		info.LockedEthHeight = hashTimer.LockedHeight
		info.EthTimerInterval = e.cfg.EthereumCfg.WithdrawInterval
		info.Amount = hashTimer.Amount.Int64()
		info.EthUserAddr = hashTimer.UserAddr
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] update [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthLockedDone))

		if err := e.withdrawCheck(int64(info.LockedEthHeight), info.EthTimerInterval, hashTimer.Amount.Int64(), hashTimer.UserAddr); err != nil {
			e.logger.Errorf("loop/withdraw check: %s [%s]", err, rHash)
			return
		}

		if err := setWithDrawNeoLockedPending(e.cfg.NEOCfg.AssetsAddress, hashTimer.UserAddr, rHash, int(info.Amount), int(e.cfg.NEOCfg.WithdrawInterval), e.store, e.neo, e.logger); err != nil {
			e.logger.Errorf("event/set WithDrawNeoLockedPending: %s [%s]", err, rHash)
			return
		}

		if err := setWithDrawNeoLockedDone(rHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, true, e.logger); err != nil {
			e.logger.Errorf("event/set WithDrawNeoLockedDone: %s [%s]", err, rHash)
			return
		}

		setWithdrawLimitExceeded(hashTimer.UserAddr)
	case eth.DestroyUnlock:
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.WithDrawEthUnlockPending {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.WithDrawEthUnlockPending), rHash)
			return
		}

		info.State = types.WithDrawEthUnlockDone
		info.UnlockedEthHash = tx
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))

	case eth.DestroyFetch: // user fetch
		info, _ = e.store.GetLockerInfo(rHash)
		if info.State != types.WithDrawNeoFetchDone {
			e.logger.Infof("[%d] locker state %s not match %s, [%s] ", state, types.LockerStateToString(info.State), types.LockerStateToString(types.WithDrawNeoFetchDone), rHash)
			return
		}

		// update info
		info.State = types.WithDrawEthFetchDone
		info.UnlockedEthHash = tx
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthFetchDone))
	}
}

func (e *EventAPI) withdrawCheck(startHeight, interval int64, amount int64, userAddr string) error {
	if b, h := e.eth.HasConfirmedBlocksHeight(startHeight, getLockDeadLineHeight(interval)); b {
		return fmt.Errorf("lock time deadline has been exceeded [%d -> %d]", startHeight, h)
	}

	if amount < e.cfg.MinWithdrawAmount {
		return fmt.Errorf("withdraw locked amount %d should more than %d", amount, e.cfg.MinWithdrawAmount)
	}

	if isWithdrawLimitExceeded(userAddr) {
		return fmt.Errorf("withdraw account %s exceed limit  ", userAddr)
	}
	return nil
}

func (e *EventAPI) loopLockerState() {
	interval := time.Duration(e.cfg.StateInterval)
	cTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-cTicker.C:
			infos := make([]*types.LockerInfo, 0)
			dInfos := make([]*types.LockerInfo, 0)
			if err := e.store.GetLockerInfos(func(info *types.LockerInfo) error {
				if info.State != types.DepositNeoUnLockedDone &&
					info.State != types.DepositNeoFetchDone &&
					info.State != types.WithDrawEthUnlockDone &&
					info.State != types.WithDrawEthFetchDone {
					infos = append(infos, info)
				} else {
					if info.Deleted == types.DeletedPending {
						dInfos = append(dInfos, info)
					}
				}
				return nil
			}); err != nil {
				e.logger.Errorf("loop/getLockerInfos: %s", err)
			}

			for _, i := range dInfos {
				if i.Interruption == true {
					continue
				}
				info, _ := e.store.GetLockerInfo(i.RHash)
				currentTime := time.Now().Unix()
				if currentTime-info.DeletedTime > 10*60 {
					e.logger.Infof("loop/check locker info deleted state [%s]", info.RHash)
					_, err := e.neo.QuerySwapInfo(info.RHash)
					if err == nil {
						continue
					}
					ht, err := e.eth.GetHashTimer(info.RHash)
					if err != nil {
						continue
					}
					if ht.LockedHeight != 0 || ht.UnlockedHeight != 0 {
						continue
					}
					info.Deleted = types.DeletedDone
					if err := e.store.UpdateLockerInfo(info); err != nil {
						e.logger.Errorf("loop/updateLocker: %s [%s]", err, info.RHash)
					} else {
						e.logger.Warnf("loop/deleted done, [%s]", info.RHash)
					}
				}
			}

			for _, i := range infos {
				if i.Interruption == true {
					continue
				}
				info, _ := e.store.GetLockerInfo(i.RHash)
				// judge if user locker is timeout, user fetch must after wrapper done
				if info.State == types.DepositNeoLockedDone || info.State == types.DepositEthFetchDone {
					if !info.NeoTimeout {
						if b, h := e.neo.HasConfirmedBlocksHeight(info.LockedNeoHeight, info.NeoTimerInterval); b {
							info.NeoTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loop/updateLocker: %s [%s]", err, info.RHash)
							} else {
								e.logger.Infof("loop/set neo timeout flag true, [%s], [%s, %d->%d]", info.RHash, types.LockerStateToString(info.State), info.LockedNeoHeight, h)
							}
						}
					}
				}

				if info.State == types.WithDrawEthLockedDone || info.State == types.WithDrawNeoFetchDone {
					if !info.EthTimeout {
						if b, h := e.eth.HasConfirmedBlocksHeight(int64(info.LockedEthHeight), info.EthTimerInterval); b {
							info.EthTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loop/updateLocker: %s [%s]", err, info.RHash)
							} else {
								e.logger.Infof("loop/set eth timeout flag true, [%s], [%s, %d->%d]", info.RHash, types.LockerStateToString(info.State), info.LockedEthHeight, h)
							}
						}
					}
				}

				switch info.State {
				case types.DepositNeoLockedDone: // check if timeout
					e.continueDepositNeoLockedDone(info.RHash)
				case types.DepositEthLockedPending: // should confirmed tx
					e.continueDepositEthLockedPending(info.RHash)
				case types.DepositEthLockedDone: // check if timeout or eth already unlock
					e.continueDepositEthLockedDone(info.RHash)
				case types.DepositEthUnLockedDone: // continue to unlock neo
					lock(info.RHash, e.logger)
					if err := setDepositNeoUnLockedPending(info.RHash, e.neo, e.store, e.cfg.NEOCfg.SignerAddress, e.logger); err != nil {
						e.logger.Errorf("loop/set DepositNeoUnLockedPending: %s [%s -> %s]", err, info.ROrigin, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.DepositEthFetchPending: // should confirmed tx
					e.continueDepositEthFetchPending(info.RHash)
				case types.DepositNeoUnLockedPending: // should confirmed tx
					lock(info.RHash, e.logger)
					e.logger.Infof("loop/continue locker state %s [%s]", types.LockerStateToString(info.State), info.RHash)
					if err := setDepositNeoUnLockedDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, false, e.logger); err != nil {
						e.logger.Errorf("loop/set DepositNeoUnLockedDone: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.DepositNeoFetchPending: // should confirm tx
					lock(info.RHash, e.logger)
					if err := setDepositNeoFetchDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, false, e.logger); err != nil {
						e.logger.Errorf("loop/set depositNeoFetchDone: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.WithDrawInit:
					e.continueWithDrawInit(info.RHash)
				case types.WithDrawNeoLockedPending: // should confirm tx
					lock(info.RHash, e.logger)
					e.logger.Infof("loop/continue locker state %s [%s]", types.LockerStateToString(info.State), info.RHash)
					if err := setWithDrawNeoLockedDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, false, e.logger); err != nil {
						e.logger.Errorf("loop/set WithDrawNeoLockedDone: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.WithDrawNeoLockedDone: // check if timeout, waiting user claim
					e.continueWithdrawNeoLockedDone(info.RHash)
				case types.WithDrawNeoUnLockedPending: // should confirm tx
					lock(info.RHash, e.logger)
					if err := setWithDrawNeoUnLockedDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, false, e.logger); err != nil {
						e.logger.Errorf("loop/set WithDrawNeoUnLockedDone: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.WithDrawNeoUnLockedDone: // wrapper should unlock on eth
					lock(info.RHash, e.logger)
					if err := setWithDrawEthUnlockPending(info.RHash, e.eth, e.store, e.cfg.EthereumCfg.OwnerAddress, e.logger); err != nil {
						e.logger.Errorf("loop/set WithDrawEthUnlockPending: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				case types.WithDrawEthUnlockPending: // should confirmed tx
					e.continueWithDrawEthUnlockPending(info.RHash)
				case types.WithDrawNeoFetchPending: // should confirmed tx
					lock(info.RHash, e.logger)
					e.logger.Infof("loop/continue locker state %s [%s]", types.LockerStateToString(info.State), info.RHash)
					if err := setWithDrawNeoFetchDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, false, e.logger); err != nil {
						e.logger.Errorf("loop/set WithDrawNeoFetchDone: %s [%s]", err, info.RHash)
					}
					unlock(info.RHash, e.logger)
				}
			}
		}
	}
}

func (e *EventAPI) continueDepositNeoLockedDone(rHash string) {
	info, _ := e.store.GetLockerInfo(rHash)
	if !info.NeoTimeout {
		if b, h := e.neo.HasConfirmedBlocksHeight(info.LockedNeoHeight, info.NeoTimerInterval); b {
			info.NeoTimeout = true
			if err := e.store.UpdateLockerInfo(info); err != nil {
				e.logger.Errorf("loop/updateLocker: %s [%s]", err, info.RHash)
			} else {
				e.logger.Infof("loop/set neo timeout flag true, [%s], [%s, %d->%d]", info.RHash, types.LockerStateToString(info.State), info.LockedNeoHeight, h)
			}
		}
	}
}

func (e *EventAPI) continueDepositEthLockedPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.DepositEthLockedDone {
		return
	}

	var hashTimer *eth.HashTimer
	hashTimer, err := e.eth.GetHashTimer(info.RHash)
	if err != nil {
		e.logger.Errorf("event/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", info.State, err, info.RHash, info.LockedEthHash)
		return
	}
	if hashTimer.LockedHeight > 0 && hashTimer.UnlockedHeight == 0 {
		e.logger.Infof("loop/continue deposit eth locked pending [%s]", info.RHash)
		info.State = types.DepositEthLockedDone
		info.LockedEthHeight = hashTimer.LockedHeight
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedDone))
	}
}

// deposit fetch, from eth
func (e *EventAPI) continueDepositEthLockedDone(rHash string) {
	info, _ := e.store.GetLockerInfo(rHash)
	if b, h := e.eth.HasConfirmedBlocksHeight(int64(info.LockedEthHeight), info.EthTimerInterval); b {
		e.logger.Infof("loop/deposit wrapper eth timeout, rHash[%s], lockerState[%s], lockerHeight[%d -> %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedEthHeight, h)
		tx, err := e.eth.WrapperFetch(info.RHash, e.cfg.EthereumCfg.OwnerAddress)
		if err != nil {
			e.logger.Errorf("loop/wrapperFetch: %s", err)
			return
		}
		e.logger.Infof("loop/deposit fetch tx(eth): %s [%s]", tx, info.RHash)
		info.EthTimeout = true
		info.UnlockedEthHash = tx
		info.State = types.DepositEthFetchPending
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthFetchPending))
	} else {
		lock(rHash, e.logger)
		defer unlock(rHash, e.logger)

		info, _ := e.store.GetLockerInfo(rHash)
		if info.State >= types.DepositEthUnLockedDone {
			return
		}

		hashTimer, err := e.eth.GetHashTimer(info.RHash)
		if err != nil {
			e.logger.Errorf("loop/getHashTimer: %s [%s]", err, rHash)
			return
		}

		if hashTimer.UnlockedHeight > 0 && hashTimer.UnlockedHeight-hashTimer.LockedHeight <= uint32(e.cfg.EthereumCfg.DepositInterval) {
			e.logger.Infof("loop/continue deposit eth locked done [%s]", info.RHash)
			info.State = types.DepositEthUnLockedDone
			info.ROrigin = hashTimer.ROrigin
			info.EthUserAddr = hashTimer.UserAddr
			//info.UnlockedEthHash = tx
			info.UnlockedEthHeight = hashTimer.UnlockedHeight
			if err = e.store.UpdateLockerInfo(info); err != nil {
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

			if err := setDepositNeoUnLockedPending(info.RHash, e.neo, e.store, e.cfg.NEOCfg.SignerAddress, e.logger); err != nil {
				e.logger.Errorf("loop/set DepositNeoUnLockedPending: %s [%s -> %s]", err, hashTimer.ROrigin, info.RHash)
				return
			}

			if err := setDepositNeoUnLockedDone(info.RHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, true, e.logger); err != nil {
				e.logger.Errorf("loop/set DepositNeoUnLockedDone: %s [%s]", err, info.RHash)
				return
			}
		}
	}
}

func (e *EventAPI) continueDepositEthFetchPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.DepositEthFetchDone {
		return
	}

	hashTimer, err := e.eth.GetHashTimer(info.RHash)
	if err != nil {
		e.logger.Errorf("event/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", info.State, err, info.RHash, info.LockedEthHash)
		return
	}
	if hashTimer.UnlockedHeight > 0 && hashTimer.UnlockedHeight-hashTimer.LockedHeight > uint32(e.cfg.EthereumCfg.DepositInterval) {
		e.logger.Infof("loop/continue withdraw eth fetch pending [%s]", info.RHash)
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		info.State = types.DepositEthFetchDone
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthFetchDone))
	}
}

func (e *EventAPI) continueWithdrawNeoLockedDone(rHash string) {
	info, _ := e.store.GetLockerInfo(rHash)
	if b, h := e.neo.HasConfirmedBlocksHeight(info.LockedNeoHeight, info.NeoTimerInterval); b {
		lock(rHash, e.logger)
		defer unlock(rHash, e.logger)
		if info.State >= types.WithDrawNeoFetchPending {
			return
		}

		e.logger.Infof("loop/withdraw neo timeout, rHash[%s], lockerState[%s], lockerHeight[%d -> %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedNeoHeight, h)
		tx, err := e.neo.RefundWrapper(info.RHash, e.cfg.NEOCfg.SignerAddress)
		if err != nil {
			e.logger.Errorf("withdrawNeoFetch(neo): %s", err)
			return
		}
		e.logger.Infof("loop/withdraw fetch tx(neo): %s [%s]", tx, info.RHash)
		info.NeoTimeout = true
		info.State = types.WithDrawNeoFetchPending
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}

		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchPending))
		if err := setWithDrawNeoFetchDone(rHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, true, e.logger); err != nil {
			e.logger.Error(err)
			return
		}
	}
}

func (e *EventAPI) continueWithDrawEthUnlockPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.WithDrawEthUnlockDone {
		return
	}

	hashTimer, err := e.eth.GetHashTimer(info.RHash)
	if err != nil {
		e.logger.Errorf("event/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", info.State, err, info.RHash, info.LockedEthHash)
		return
	}
	if hashTimer.UnlockedHeight > 0 && hashTimer.UnlockedHeight-hashTimer.LockedHeight <= uint32(e.cfg.EthereumCfg.WithdrawInterval) {
		e.logger.Infof("loop/continue withdraw eth unlocked pending [%s]", info.RHash)
		info.UnlockedEthHeight = hashTimer.UnlockedHeight
		info.State = types.WithDrawEthUnlockDone
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))
	}
}

func (e *EventAPI) continueWithDrawInit(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.WithDrawEthLockedDone {
		return
	}

	hashTimer, err := e.eth.GetHashTimer(rHash)
	if err != nil {
		return
	}

	if hashTimer.LockedHeight > 0 && hashTimer.UnlockedHeight == 0 {
		e.logger.Infof("loop/continue withdraw init [%s]", info.RHash)

		info.State = types.WithDrawEthLockedDone
		info.RHash = rHash
		info.LockedEthHeight = hashTimer.LockedHeight
		info.EthTimerInterval = e.cfg.EthereumCfg.WithdrawInterval
		info.Amount = hashTimer.Amount.Int64()
		info.EthUserAddr = hashTimer.UserAddr
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("update [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthLockedDone))

		if err := e.withdrawCheck(int64(info.LockedEthHeight), info.EthTimerInterval, hashTimer.Amount.Int64(), hashTimer.UserAddr); err != nil {
			e.logger.Error("withdraw check: %s [%s]", err, rHash)
			return
		}

		if err := setWithDrawNeoLockedPending(e.cfg.NEOCfg.AssetsAddress, hashTimer.UserAddr, rHash, int(info.Amount), int(e.cfg.NEOCfg.WithdrawInterval), e.store, e.neo, e.logger); err != nil {
			e.logger.Errorf("set WithDrawNeoLockedPending: %s [%s]", err, rHash)
			return
		}

		if err := setWithDrawNeoLockedDone(rHash, e.neo, e.store, e.cfg.NEOCfg.ConfirmedHeight, true, e.logger); err != nil {
			e.logger.Errorf("set WithDrawNeoLockedDone: %s [%s]", err, rHash)
			return
		}
		setWithdrawLimitExceeded(hashTimer.UserAddr)
	}
}

func getLockDeadLineHeight(height int64) int64 {
	return (height / 2) - 5
}

func setDepositNeoUnLockedDone(rHash string, neoTransaction *neo.Transaction, s *store.Store, confirmedHeight int, sync bool, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.DepositNeoUnLockedDone
	if info.State >= targetState {
		return nil
	}

	if !sync {
		if _, err := neoTransaction.QuerySwapInfo(rHash); err != nil {
			return nil
		}
	}

	swapInfo, err := neoTransaction.QuerySwapInfoAndConfirmedTx(rHash, neo.WrapperUnlock, confirmedHeight)
	if err != nil {
		return err
	}

	info.State = targetState
	info.UnlockedNeoHeight = swapInfo.UnlockedHeight
	info.UnlockedNeoHash = swapInfo.TxIdOut

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setDepositNeoFetchDone(rHash string, neoTransaction *neo.Transaction, s *store.Store, confirmedHeight int, sync bool, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.DepositNeoFetchDone
	if info.State >= targetState {
		return nil
	}

	if !sync {
		if _, err := neoTransaction.QuerySwapInfo(rHash); err != nil {
			return nil
		}
	}

	swapInfo, err := neoTransaction.QuerySwapInfoAndConfirmedTx(rHash, neo.RefundUser, confirmedHeight)
	if err != nil {
		return err
	}

	info.State = targetState
	info.UnlockedNeoHash = swapInfo.TxIdRefund
	info.UnlockedNeoHeight = swapInfo.UnlockedHeight

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setDepositNeoUnLockedPending(rHash string, neoTransaction *neo.Transaction, s *store.Store, signerAddress string, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.DepositNeoUnLockedPending
	if info.State >= targetState {
		return nil
	}
	txHash, err := neoTransaction.WrapperUnlock(info.ROrigin, signerAddress, info.EthUserAddr)
	if err != nil {
		return err
	}
	logger.Infof("deposit/wrapper unlock(neo): %s [%s] ", txHash, rHash)
	info.State = targetState

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setWithDrawNeoFetchDone(rHash string, neoTransaction *neo.Transaction, s *store.Store, confirmedHeight int, sync bool, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.WithDrawNeoFetchDone
	if info.State >= targetState {
		return nil
	}

	if !sync {
		if _, err := neoTransaction.QuerySwapInfo(rHash); err != nil {
			return nil
		}
	}

	swapInfo, err := neoTransaction.QuerySwapInfoAndConfirmedTx(rHash, neo.RefundWrapper, confirmedHeight)
	if err != nil {
		return err
	}
	info.State = targetState
	info.UnlockedNeoHeight = swapInfo.UnlockedHeight
	info.UnlockedNeoHash = swapInfo.TxIdRefund

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setWithDrawNeoUnLockedDone(rHash string, neoTransaction *neo.Transaction, s *store.Store, confirmedHeight int, sync bool, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}

	targetState := types.WithDrawNeoUnLockedDone
	if info.State >= targetState {
		return nil
	}

	if !sync {
		if _, err := neoTransaction.QuerySwapInfo(rHash); err != nil {
			return nil
		}
	}

	swapInfo, err := neoTransaction.QuerySwapInfoAndConfirmedTx(rHash, neo.UserUnlock, confirmedHeight)
	if err != nil {
		return err
	}
	logger.Infof("swap info: %s", hubUtil.ToString(swapInfo))
	info.State = targetState
	info.UnlockedNeoHash = swapInfo.TxIdOut
	info.ROrigin = swapInfo.OriginText
	info.NeoUserAddr = swapInfo.UserNeoAddress
	info.UnlockedNeoHeight = swapInfo.UnlockedHeight

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setWithDrawNeoLockedPending(assetsAddress, userEthAddr, rHash string, amount, withdrawInterval int, s *store.Store, neoTransaction *neo.Transaction, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.WithDrawNeoLockedPending
	if info.State >= targetState {
		return nil
	}

	txHash, err := neoTransaction.WrapperLock(assetsAddress, userEthAddr, rHash, amount, withdrawInterval)
	if err != nil {
		return fmt.Errorf("event/wrapper lock(neo): %s [%s]", err, rHash)
	}
	logger.Infof("withdraw/wrapper neo lock tx: %s [%s]", txHash, info.RHash)
	info.State = targetState
	if err = s.UpdateLockerInfo(info); err != nil {
		return err
	}
	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return nil
}

func setWithDrawNeoLockedDone(rHash string, neoTransaction *neo.Transaction, s *store.Store, confirmedHeight int, sync bool, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.WithDrawNeoLockedDone
	if info.State >= targetState {
		return nil
	}

	if !sync {
		if _, err := neoTransaction.QuerySwapInfo(rHash); err != nil {
			return nil
		}
	}

	swapInfo, err := neoTransaction.QuerySwapInfoAndConfirmedTx(rHash, neo.WrapperLock, confirmedHeight)
	if err != nil {
		return err
	}

	info.State = targetState
	info.LockedNeoHeight = swapInfo.LockedHeight
	info.LockedNeoHash = swapInfo.TxIdIn
	info.NeoTimerInterval = swapInfo.OvertimeBlocks

	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

func setWithDrawEthUnlockPending(rHash string, ethTransaction *eth.Transaction, s *store.Store, signerAddr string, logger *zap.SugaredLogger) error {
	info, err := s.GetLockerInfo(rHash)
	if err != nil {
		return err
	}
	targetState := types.WithDrawEthUnlockPending
	if info.State >= targetState {
		return nil
	}

	tx, gasPrice, err := ethTransaction.WrapperUnlock(info.RHash, info.ROrigin, signerAddr)
	if err != nil {
		return err
	}
	logger.Infof("withdraw/wrapper eth unlock: %s [%s]", tx, info.RHash)
	info.State = targetState
	info.GasPrice = gasPrice
	info.UnlockedEthHash = tx
	logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(targetState))
	return s.UpdateLockerInfo(info)
}

var withdrawTimeLimit = new(sync.Map)
var logger = log.NewLogger("timelimit")

func resetWithdrawTimeLimit(ctx context.Context, interval int) {
	cTicker := time.NewTicker(time.Duration(interval) * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-cTicker.C:
			withdrawTimeLimit.Range(func(key, value interface{}) bool {
				withdrawTimeLimit.Store(key, false)
				return true
			})
		}
	}
}

func isWithdrawLimitExceeded(addr string) bool {
	if r, ok := withdrawTimeLimit.Load(RemovePrefix(addr)); ok {
		return r.(bool)
	} else {
		return false
	}
}

func setWithdrawLimitExceeded(addr string) {
	logger.Info("==== set ", RemovePrefix(addr))
	withdrawTimeLimit.Store(RemovePrefix(addr), true)
}

func RemovePrefix(str string) string {
	if len(str) == 42 && strings.HasPrefix(str, "0x") {
		return str[2:]
	}
	return str
}
