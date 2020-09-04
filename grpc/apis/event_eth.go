package apis

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func (e *EventAPI) ethEventLister() {
	contractAddress := common.HexToAddress(e.cfg.EthereumCfg.Contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	filterer, err := eth.NewQLCChainFilterer(contractAddress, e.eth)
	if err != nil {
		e.logger.Error("NewQLCChainFilterer: ", err)
		return
	}
	logs := make(chan ethTypes.Log)
	sub, err := e.eth.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		e.logger.Error("SubscribeFilterLogs: ", err)
		return
	}
	for {
		select {
		case <-e.ctx.Done():
			return
		case err := <-sub.Err():
			e.logger.Error("SubscribeFilterLogs: ", err)
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
			go e.processEthEvent(state, rHash, txHash, txHeight)
		}
	}
}

func (e *EventAPI) processEthEvent(state int64, rHash, tx string, txHeight uint64) {
	var info *types.LockerInfo
	var err error
	defer func() {
		e.store.SetLockerStateFail(info, err)
	}()

	var b bool
	e.logger.Infof("waiting for eth tx %s confirmed ", tx)
	if b, err = eth.TxVerifyAndConfirmed(tx, int64(txHeight), int64(e.cfg.EthereumCfg.ConfirmedHeight), e.eth); !b || err != nil {
		e.logger.Errorf("event/txVerify(eth)[%d]: %s, %v, rHash[%s], txHash[%s]", state, err, b, rHash, tx)
		return
	}

	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)

	if eth.State(state) != eth.DestroyLock {
		if info, err = e.store.GetLockerInfo(rHash); err != nil {
			e.logger.Errorf("event/getLockerInfo[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
			return
		}
	}

	var hashTimer *eth.HashTimer
	if hashTimer, err = eth.GetHashTimer(e.eth, e.cfg.EthereumCfg.Contract, rHash); err != nil {
		e.logger.Errorf("event/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
		return
	}

	var txHash string
	switch eth.State(state) {
	case eth.IssueLock:
		info.State = types.DepositEthLockedDone
		info.LockedErc20Hash = tx
		info.LockedErc20Height = hashTimer.LockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthLockedDone))

	case eth.IssueUnlock:
		info.State = types.DepositEthUnLockedDone
		info.ROrigin = hashTimer.ROrigin
		info.UnlockedErc20Hash = tx
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

		// to neo unlock
		if txHash, err = e.neo.WrapperUnlock(hashTimer.ROrigin, e.cfg.NEOCfg.WIF, hashTimer.UserAddr); err != nil {
			e.logger.Errorf("ethEvent/wrapperUnlock[%d]: %s, %s, %s, [%s]", state, err, hashTimer.ROrigin, hashTimer.UserAddr, rHash)
			return
		}
		e.logger.Infof("[%d] deposit/wrapper unlock(neo): %s [%s] ", state, txHash, rHash)
		info.State = types.DepositNeoUnLockedPending
		info.UnlockedNep5Hash = txHash
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositNeoUnLockedPending))

		var height uint32
		e.logger.Infof("waiting for neo tx %s confirmed", txHash)
		if b, height, err = e.neo.TxVerifyAndConfirmed(txHash, e.cfg.NEOCfg.ConfirmedHeight); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo)[%d]: %s, %v, %s, [%s]", state, err, b, txHash, rHash)
			return
		}
		info.State = types.DepositNeoUnLockedDone
		info.UnlockedNep5Height = height
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositNeoUnLockedDone))

	case eth.IssueFetch: // wrapper Fetch
		info.State = types.DepositEthFetchDone
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthFetchDone))

	case eth.DestroyLock:
		info := new(types.LockerInfo)
		info.State = types.WithDrawEthLockedDone
		info.RHash = rHash
		info.LockedErc20Hash = tx
		info.LockedErc20Height = hashTimer.LockedHeight
		info.Amount = hashTimer.Amount.Int64()
		info.Erc20Addr = hashTimer.UserAddr
		if err = e.store.AddLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] add [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthLockedDone))

		// neo lock
		if b, h := eth.IsBeyondIntervalHeight(int64(info.LockedErc20Height), getLockDeadLineHeight(e.cfg.EthereumCfg.WithdrawHeight), e.eth); b {
			err = fmt.Errorf("lock time deadline has been exceeded [%s] [%d -> %d]", info.RHash, info.LockedErc20Height, h)
			e.logger.Error(err)
			return
		}

		txHash, err = e.neo.WrapperLock(e.cfg.NEOCfg.WIF, hashTimer.UserAddr, rHash, int(info.Amount))
		if err != nil {
			e.logger.Errorf("ethEvent/wrapper lock(neo)[%d]: %s [%s]", state, err, rHash)
			return
		}
		e.logger.Infof("[%d] withdraw/wrapper neo lock tx: %s [%s]", state, txHash, info.RHash)
		info.State = types.WithDrawNeoLockedPending
		info.LockedNep5Hash = txHash
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedPending))

		var height uint32
		e.logger.Infof("waiting for neo tx %s confirmed", txHash)
		if b, height, err = e.neo.TxVerifyAndConfirmed(txHash, e.cfg.NEOCfg.ConfirmedHeight); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo)[%d]: %s, %v, %s [%s]", state, err, b, txHash, rHash)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.LockedNep5Height = height
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedDone))

	case eth.DestroyUnlock:
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		info.State = types.WithDrawEthUnlockDone
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))

	case eth.DestroyFetch: // user fetch
		// update info
		info.State = types.WithDrawEthFetchDone
		info.UnlockedErc20Hash = tx
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] set [%s] state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthFetchDone))
	}
}

func (e *EventAPI) loopLockerState() {
	cTicker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-cTicker.C:
			infos := make([]*types.LockerInfo, 0)
			if err := e.store.GetLockerInfos(func(info *types.LockerInfo) error {
				infos = append(infos, info)
				return nil
			}); err != nil {
				e.logger.Errorf("loopLockerState/getLockerInfos: %s", err)
			}
			for _, info := range infos {
				// user timeout -> fetch neo
				if info.State >= types.DepositNeoLockedDone && info.State <= types.DepositEthFetchDone {
					if !info.NeoTimeout && info.State == types.DepositEthFetchDone {
						if b, h := e.neo.IsBeyondIntervalHeight(info.LockedNep5Height, e.cfg.NEOCfg.DepositHeight); b {
							info.NeoTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loopLockerState/updateLocker: %s [%s]", err, info.RHash)
							}
							e.logger.Infof("[%s] set neo timeout flag true, [%s, %d->%d]", info.RHash, types.LockerStateToString(info.State), info.LockedNep5Height, h)
						}
					}
				}

				if info.State >= types.WithDrawEthLockedDone && info.State <= types.WithDrawNeoFetchDone {
					if !info.EthTimeout && info.State == types.WithDrawNeoFetchDone {
						if b, h := eth.IsBeyondIntervalHeight(int64(info.LockedErc20Height), e.cfg.EthereumCfg.WithdrawHeight, e.eth); b {
							info.EthTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loopLockerState/updateLocker: %s [%s]", err, info.RHash)
							}
							e.logger.Infof("[%s] set eth timeout flag true, [%s, %d->%d]", info.RHash, types.LockerStateToString(info.State), info.LockedErc20Height, h)
						}
					}
				}

				switch info.State {
				case types.DepositEthLockedPending: // should confirmed tx
					e.continueDepositEthLockedPending(info.RHash)
				case types.DepositEthLockedDone: // check if timeout or eth already unlock
					e.continueDepositEthLockedDone(info.RHash)
				case types.DepositNeoUnLockedPending: // should confirmed tx
					e.continueDepositNeoUnLockedPending(info.RHash)
				case types.WithDrawNeoLockedDone: // check if timeout or neo already unlocked
					e.continueWithdrawNeoLockedDone(info.RHash)
				case types.WithDrawNeoUnLockedDone: // wrapper should unlock on eth
					e.continueWithDrawNeoUnLockedDone(info.RHash)
				case types.WithDrawEthUnlockPending: // should confirmed tx
					e.continueWithDrawEthUnlockPending(info.RHash)
				case types.WithDrawNeoFetchPending: // should confirmed tx
					e.continueWithDrawNeoFetchPending(info.RHash)
				}
			}
		}
	}
}

func (e *EventAPI) continueDepositNeoUnLockedPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.DepositNeoUnLockedDone {
		return
	}
	e.logger.Infof("loop/continue deposit neo unlocked pending : %s", info.RHash)

	b, height, err := e.neo.TxVerifyAndConfirmed(info.UnlockedNep5Hash, e.cfg.NEOCfg.ConfirmedHeight)
	if !b || err != nil {
		e.logger.Errorf("loop/txVerify(neo)[%d]: %s, %v, %s, [%s]", info.State, err, b, info.UnlockedNep5Hash, info.RHash)
		return
	}
	info.State = types.DepositNeoUnLockedDone
	info.UnlockedNep5Height = height
	if err = e.store.UpdateLockerInfo(info); err != nil {
		return
	}
	e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoUnLockedDone))
}

func (e *EventAPI) continueDepositEthLockedPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.DepositEthLockedDone {
		return
	}

	e.logger.Infof("loop/continue deposit eth locked pending : %s", info.RHash)
	var hashTimer *eth.HashTimer
	hashTimer, err := eth.GetHashTimer(e.eth, e.cfg.EthereumCfg.Contract, info.RHash)
	if err != nil {
		e.logger.Errorf("ethEvent/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", info.State, err, info.RHash, info.LockedErc20Height)
		return
	}
	if hashTimer.LockedHeight > 0 {
		info.State = types.DepositEthLockedDone
		info.LockedErc20Height = hashTimer.LockedHeight
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedDone))
	}
}

// deposit fetch, from eth
func (e *EventAPI) continueDepositEthLockedDone(rHash string) {
	info, _ := e.store.GetLockerInfo(rHash)
	if b, h := eth.IsBeyondIntervalHeight(int64(info.LockedErc20Height), e.cfg.EthereumCfg.DepositHeight, e.eth); b {
		e.logger.Infof("loop/deposit wrapper eth timeout, rHash[%s], lockerState[%s], lockerHeight[%d -> %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedErc20Height, h)
		tx, err := eth.WrapperFetch(info.RHash, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.eth)
		if err != nil {
			e.logger.Errorf("loop/wrapperFetch: %s", err)
			return
		}
		e.logger.Infof("loop/deposit fetch tx(eth): %s [%s]", tx, info.RHash)
		info.EthTimeout = true
		info.UnlockedErc20Hash = tx
		info.State = types.DepositEthFetchPending
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
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

		hashTimer, err := eth.GetHashTimer(e.eth, e.cfg.EthereumCfg.Contract, info.RHash)
		if err != nil {
			e.logger.Errorf("loop/getHashTimer: %s [%s]", err, rHash)
			return
		}

		if hashTimer.UnlockedHeight > 0 {
			info.State = types.DepositEthUnLockedDone
			info.ROrigin = hashTimer.ROrigin
			info.UnlockedErc20Height = hashTimer.UnlockedHeight
			if err = e.store.UpdateLockerInfo(info); err != nil {
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

			// to neo unlock
			txHash, err := e.neo.WrapperUnlock(hashTimer.ROrigin, e.cfg.NEOCfg.WIF, hashTimer.UserAddr)
			if err != nil {
				e.logger.Errorf("loop/wrapperUnlock: %s, %s, %s, [%s]", err, hashTimer.ROrigin, hashTimer.UserAddr, rHash)
				return
			}
			e.logger.Infof("loop/deposit wrapper unlock(neo): %s [%s] ", txHash, rHash)
			info.State = types.DepositNeoUnLockedPending
			info.UnlockedNep5Hash = txHash
			if err = e.store.UpdateLockerInfo(info); err != nil {
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoUnLockedPending))

			var height uint32
			e.logger.Infof("waiting for neo tx %s confirmed", txHash)
			b, height, err := e.neo.TxVerifyAndConfirmed(txHash, e.cfg.NEOCfg.ConfirmedHeight)
			if !b || err != nil {
				e.logger.Errorf("loop/txVerify(neo): %s, %v, %s, [%s]", err, b, txHash, rHash)
				return
			}
			info.State = types.DepositNeoUnLockedDone
			info.UnlockedNep5Height = height
			if err = e.store.UpdateLockerInfo(info); err != nil {
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoUnLockedDone))
		}
	}
}

// withdraw fetch, from neo
func (e *EventAPI) continueWithdrawNeoLockedDone(rHash string) {
	info, _ := e.store.GetLockerInfo(rHash)
	if b, h := e.neo.IsBeyondIntervalHeight(info.LockedNep5Height, e.cfg.NEOCfg.WithdrawHeight); b {
		e.logger.Infof("loop/withdraw neo timeout, rHash[%s], lockerState[%s], lockerHeight[%d -> %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedNep5Height, h)
		tx, err := e.neo.RefundWrapper(info.RHash, e.cfg.NEOCfg.WIF)
		if err != nil {
			e.logger.Errorf("withdrawNeoFetch(neo): %s", err)
			return
		}
		e.logger.Infof("loop/withdraw fetch tx(neo): %s [%s]", tx, info.RHash)
		info.NeoTimeout = true
		info.State = types.WithDrawNeoFetchPending
		info.UnlockedNep5Hash = tx
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchPending))
		e.logger.Infof("loop/waiting for neo tx %s confirmed", tx)
		b, height, err := e.neo.TxVerifyAndConfirmed(tx, e.cfg.NEOCfg.ConfirmedHeight)
		if !b || err != nil {
			e.logger.Errorf("txVerify(neo): %s, %v, %s, [%s]", err, b, tx, info.RHash)
			return
		}
		info.State = types.WithDrawNeoFetchDone
		info.UnlockedNep5Height = height
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchDone))
	} else {
		lock(rHash, e.logger)
		defer unlock(rHash, e.logger)
		info, _ := e.store.GetLockerInfo(rHash)
		if info.State >= types.WithDrawNeoUnLockedDone {
			return
		}
		e.logger.Infof("loop/continue withdraw neo locked done : %s", info.RHash)
		swapInfo, err := e.neo.QuerySwapInfo(info.RHash)
		if err != nil {
			e.logger.Error(err)
			return
		}
		if swapInfo.State == 5 { //todo need confirmed
			e.logger.Infof("loop/continue withdraw neo locked done  [%s]", info.RHash)
			e.logger.Infof("swap info : %s", util.ToString(swapInfo))

			rOrigin := swapInfo.OriginText
			unlockedTxHash := swapInfo.OriginText //todo

			b, height, err := e.neo.TxVerifyAndConfirmed(unlockedTxHash, e.cfg.NEOCfg.ConfirmedHeight)
			if !b || err != nil {
				e.logger.Errorf("neo tx confirmed: %s, %v, %s [%s]", err, b, unlockedTxHash, info.RHash)
				return
			}

			info.State = types.WithDrawNeoUnLockedDone
			info.UnlockedNep5Height = height
			info.UnlockedNep5Hash = unlockedTxHash
			info.ROrigin = rOrigin
			if err := e.store.UpdateLockerInfo(info); err != nil {
				e.logger.Errorf("%s: %s", info.RHash, err)
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoUnLockedDone))

			tx, err := eth.WrapperUnlock(info.RHash, rOrigin, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.eth)
			if err != nil {
				e.logger.Errorf("eth wrapper unlock: %s [%s]", err, info.RHash)
				return
			}
			e.logger.Infof("loop/withdraw wrapper eth unlock: %s [%s] ", tx, info.RHash)
			info.State = types.WithDrawEthUnlockPending
			info.UnlockedErc20Hash = tx
			if err := e.store.UpdateLockerInfo(info); err != nil {
				e.logger.Errorf("%s: %s", info.RHash, err)
				return
			}
			e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
		}
	}
}

func (e *EventAPI) continueWithDrawNeoUnLockedDone(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.WithDrawEthUnlockPending {
		return
	}
	e.logger.Infof("loop/continue withdraw neo unlocked done : %s", info.RHash)
	tx, err := eth.WrapperUnlock(info.RHash, info.ROrigin, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.eth)
	if err != nil {
		e.logger.Errorf("eth wrapper unlock: %s [%s]", err, info.RHash)
		return
	}
	e.logger.Infof("loop/withdraw wrapper eth unlock: %s [%s]", tx, info.RHash)
	info.State = types.WithDrawEthUnlockPending
	info.UnlockedErc20Hash = tx
	if err := e.store.UpdateLockerInfo(info); err != nil {
		e.logger.Errorf("%s: %s", info.RHash, err)
		return
	}
	e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockPending))
}

func (e *EventAPI) continueWithDrawEthUnlockPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.WithDrawEthUnlockDone {
		return
	}
	e.logger.Infof("loop/continue withdraw eth unlocked pending : %s", info.RHash)

	hashTimer, err := eth.GetHashTimer(e.eth, e.cfg.EthereumCfg.Contract, info.RHash)
	if err != nil {
		e.logger.Errorf("ethEvent/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", info.State, err, info.RHash, info.LockedErc20Height)
		return
	}
	if hashTimer.UnlockedHeight > 0 {
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		info.State = types.WithDrawEthUnlockDone
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))
	}
}

func (e *EventAPI) continueWithDrawNeoFetchPending(rHash string) {
	lock(rHash, e.logger)
	defer unlock(rHash, e.logger)
	info, _ := e.store.GetLockerInfo(rHash)
	if info.State >= types.WithDrawNeoFetchDone {
		return
	}

	e.logger.Infof("loop/continue withdraw neo fetch pending  [%s]", info.RHash)
	e.logger.Infof("waiting for neo tx %s confirmed", info.UnlockedNep5Hash)
	b, height, err := e.neo.TxVerifyAndConfirmed(info.UnlockedNep5Hash, e.cfg.NEOCfg.ConfirmedHeight)
	if !b || err != nil {
		e.logger.Errorf("txVerify(neo): %s, %v, %s, [%s]", err, b, info.UnlockedNep5Hash, info.RHash)
		return
	}
	info.State = types.WithDrawNeoFetchDone
	info.UnlockedNep5Height = height
	if err := e.store.UpdateLockerInfo(info); err != nil {
		e.logger.Error(err)
		return
	}
	e.logger.Infof("loop/set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchDone))
}

func getLockDeadLineHeight(height int64) int64 {
	return (height / 2) - 5
}
