package apis

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/types"
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
		if info != nil && err != nil {
			info.Remark = err.Error()
			e.store.SetLockerStateFail(info, err)
		}
	}()

	if eth.State(state) != eth.DestroyLock {
		if info, err = e.store.GetLockerInfo(rHash); err != nil {
			e.logger.Errorf("ethEvent/getLockerInfo[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
			return
		}
	}

	var b bool
	e.logger.Infof("waiting for eth tx [%s] confirmed ", tx)
	if b, err = eth.TxVerifyAndConfirmed(tx, int64(txHeight), int64(e.cfg.EthereumCfg.ConfirmedHeight), e.eth); !b || err != nil {
		e.logger.Errorf("ethEvent/txVerify(eth)[%d]: %s, %v, rHash[%s], txHash[%s]", state, err, b, rHash, tx)
		return
	}

	var hashTimer *eth.HashTimer
	if hashTimer, err = eth.GetHashTimer(e.eth, e.cfg.EthereumCfg.Contract, rHash); err != nil {
		e.logger.Errorf("ethEvent/getHashTimer[%d]: %s, rHash[%s], txHash[%s]", state, err, rHash, tx)
		return
	}

	var txHash string
	switch eth.State(state) {
	case eth.IssueLock:
		// update indo
		info.State = types.DepositEthLockedDone
		info.LockedErc20Hash = tx
		info.LockedErc20Height = hashTimer.LockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthLockedDone))

	case eth.IssueUnlock:
		info.State = types.DepositEthUnLockedDone
		info.ROrigin = hashTimer.ROrigin
		info.UnlockedErc20Hash = tx
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

		// to neo unlock
		if txHash, err = e.neo.WrapperUnlock(hashTimer.ROrigin, e.cfg.NEOCfg.WIF, hashTimer.UserAddr); err != nil {
			e.logger.Errorf("ethEvent/wrapperUnlock[%d]: %s, %s, %s, [%s]", state, err, hashTimer.ROrigin, hashTimer.UserAddr, rHash)
			return
		}
		e.logger.Infof("[%d] [%s] deposit/wrapper unlock(neo): %s", state, rHash, txHash)
		info.State = types.DepositNeoUnLockedPending
		info.UnlockedNep5Hash = txHash
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositNeoUnLockedPending))

		var height uint32
		e.logger.Infof("waiting for neo tx [%s] confirmed", txHash)
		if b, height, err = e.neo.TxVerifyAndConfirmed(txHash, e.cfg.NEOCfg.ConfirmedHeight); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo)[%d]: %s, %v [%s]", state, err, b, rHash)
			return
		}
		info.State = types.DepositNeoUnLockedDone
		info.UnlockedNep5Height = height
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositNeoUnLockedDone))

	case eth.IssueFetch: // wrapper Fetch
		info.State = types.DepositEthFetchDone
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthFetchDone))

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
		e.logger.Infof("[%d] [%s] add state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthLockedDone))

		// neo lock
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
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedPending))

		var height uint32
		e.logger.Infof("waiting for neo tx [%s] confirmed", txHash)
		if b, height, err = e.neo.TxVerifyAndConfirmed(txHash, e.cfg.NEOCfg.ConfirmedHeight); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo)[%d]: %s, %v [%s]", state, err, b, rHash)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.LockedNep5Height = height
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedDone))

	case eth.DestroyUnlock:
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		info.State = types.WithDrawEthUnlockDone
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))

	case eth.DestroyFetch: // user fetch
		// update info
		info.State = types.WithDrawEthFetchDone
		info.UnlockedErc20Hash = tx
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawEthFetchDone))
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
						if e.neo.IsLockerTimeout(info.LockedNep5Height, e.cfg.NEOCfg.DepositHeight) {
							info.NeoTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loopLockerState/updateLocker: %s [%s]", err, info.RHash)
							}
							e.logger.Infof("[%s] set neo timeout flag true, [%s]", info.RHash, types.LockerStateToString(info.State))
						}
					}
				}

				if info.State >= types.WithDrawEthLockedDone && info.State <= types.WithDrawNeoFetchDone {
					if !info.EthTimeout && info.State == types.WithDrawNeoFetchDone {
						if eth.IsLockerTimeout(int64(info.LockedErc20Height), e.cfg.EthereumCfg.WithdrawHeight, e.eth) {
							info.EthTimeout = true
							if err := e.store.UpdateLockerInfo(info); err != nil {
								e.logger.Errorf("loopLockerState/updateLocker: %s [%s]", err, info.RHash)
							}
							e.logger.Infof("[%s] set eth timeout flag true, [%s, %d]", info.RHash, types.LockerStateToString(info.State), e.cfg.EthereumCfg.WithdrawHeight)
						}
					}
				}

				switch info.State {
				case types.DepositEthLockedDone:
					e.depositEthFetch(info)
				case types.WithDrawNeoLockedDone:
					e.withdrawNeoFetch(info)
				case types.WithDrawNeoUnLockedDone:
					//todo add locker
					//e.withdrawNeoFetch(info)
				}
			}
		}
	}
}

// deposit fetch, from eth
func (e *EventAPI) depositEthFetch(info *types.LockerInfo) {
	if eth.IsLockerTimeout(int64(info.LockedErc20Height), e.cfg.EthereumCfg.DepositHeight, e.eth) {
		e.logger.Infof("deposit/wrapper eth timeout, rHash[%s], lockerState[%s], lockerHeight[%d, %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedErc20Height, e.cfg.EthereumCfg.DepositHeight)
		tx, err := eth.WrapperFetch(info.RHash, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.eth)
		if err != nil {
			e.logger.Errorf("depositEthFetch/wrapperFetch: %s", err)
			return
		}
		e.logger.Infof("deposit/wrapper fetch tx(eth): %s [%s]", tx, info.RHash)
		info.EthTimeout = true
		info.UnlockedErc20Hash = tx
		info.State = types.DepositEthFetchPending
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthFetchPending))
	}
}

// withdraw fetch, from neo
func (e *EventAPI) withdrawNeoFetch(info *types.LockerInfo) {
	if e.neo.IsLockerTimeout(info.LockedNep5Height, e.cfg.NEOCfg.WithdrawHeight) {
		e.logger.Infof("withdraw/wrapper neo timeout, rHash[%s], lockerState[%s], lockerHeight[%d, %d]", info.RHash,
			types.LockerStateToString(info.State), info.LockedNep5Height, e.cfg.NEOCfg.WithdrawHeight)
		tx, err := e.neo.RefundWrapper(info.RHash, e.cfg.NEOCfg.WIF)
		if err != nil {
			e.logger.Errorf("loopLockerState/withdrawNeoFetch(neo): %s", err)
			return
		}
		e.logger.Infof("withdraw/wrapper fetch tx(neo): %s [%s]", tx, info.RHash)
		info.NeoTimeout = true
		info.State = types.WithDrawNeoFetchPending
		info.UnlockedNep5Hash = tx
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchPending))
		e.logger.Infof("waiting for neo tx [%s] confirmed", tx)
		b, height, err := e.neo.TxVerifyAndConfirmed(tx, e.cfg.NEOCfg.ConfirmedHeight)
		if !b || err != nil {
			e.logger.Errorf("loopLockerState/txVerify(neo): %s, %v [%s]", err, b, info.RHash)
			return
		}
		info.State = types.WithDrawNeoFetchDone
		info.UnlockedNep5Height = height
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchDone))
	}
}
