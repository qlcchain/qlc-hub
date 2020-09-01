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
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

func (e *EventAPI) ethEventLister() {
	contractAddress := common.HexToAddress(e.ethContract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	filterer, err := eth.NewQLCChainFilterer(contractAddress, e.ethClient)
	if err != nil {
		e.logger.Error("NewQLCChainFilterer: ", err)
		return
	}
	logs := make(chan ethTypes.Log)
	sub, err := e.ethClient.SubscribeFilterLogs(context.Background(), query, logs)
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
			e.logger.Infof("event log: rHash[%s], state[%d], txHash[%s], txHeight[%d]", rHash, state, txHash, txHeight)
			go e.processEthEvent(state, rHash, txHash, txHeight)
		}
	}
}

func (e *EventAPI) processEthEvent(state int64, rHash, tx string, txHeight uint64) {
	var info *types.LockerInfo
	var err error
	if eth.State(state) != eth.DestroyLock {
		if info, err = e.store.GetLockerInfo(rHash); err != nil {
			e.logger.Errorf("ethEvent/getLockerInfo: %s, rHash[%s], state[%d], txHash[%s]", err, rHash, state, tx)
			return
		}
	}

	var b bool
	if b, err = eth.TxVerifyAndConfirmed(tx, int64(txHeight), int64(ethConfirmedHeight), e.ethClient); !b || err != nil {
		e.logger.Errorf("ethEvent/txVerify(eth): %s, %v, rHash[%s], txHash[%s]", err, b, rHash, tx)
		return
	}

	var hashTimer *eth.HashTimer
	if hashTimer, err = eth.GetHashTimer(e.ethClient, e.ethContract, rHash); err != nil {
		e.logger.Errorf("ethEvent/getHashTimer: %s, rHash[%s], txHash[%s]", err, rHash, tx)
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

		//todo notify

	case eth.IssueUnlock:
		info.State = types.DepositEthUnLockedDone
		info.ROrigin = hashTimer.ROrigin
		info.UnlockedErc20Hash = tx
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

		//todo notify

		// to neo unlock
		if txHash, err = neo.WrapperUnlock(hashTimer.ROrigin, e.cfg.NEOCfg.WIF, hashTimer.UserAddr, e.neoTransaction); err != nil {
			e.logger.Errorf("ethEvent/wrapperUnlock: %s [%s]", err, rHash)
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
		if b, height, err = neo.TxVerifyAndConfirmed(txHash, neoConfirmedHeight, e.neoTransaction); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo): %s, %v [%s]", err, b, rHash)
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
		fmt.Println("================ ", hashTimer.String())

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
		txHash, err = neo.WrapperLock(e.cfg.NEOCfg.WIF, hashTimer.UserAddr, rHash, int(info.Amount), e.neoTransaction)
		if err != nil {
			e.logger.Errorf("ethEvent/wrapper lock(neo): %s [%s]", err, rHash)
			return
		}
		e.logger.Infof("[%d] [%s] withdraw/wrapper neo lock tx: [%s]", state, info.RHash, txHash)
		info.State = types.WithDrawNeoLockedPending
		info.LockedNep5Hash = txHash
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedPending))

		var height uint32
		if b, height, err = neo.TxVerifyAndConfirmed(txHash, neoConfirmedHeight, e.neoTransaction); !b || err != nil {
			e.logger.Errorf("ethEvent/txVerify(neo): %s, %v [%s]", err, b, rHash)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.LockedNep5Height = height
		if err = e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("[%d] [%s] set state to [%s]", state, info.RHash, types.LockerStateToString(types.WithDrawNeoLockedDone))
		//todo notify

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
				switch info.State {
				case types.DepositEthLockedDone:
					e.depositFetch(info)
				case types.WithDrawNeoLockedDone:
					e.withdrawFetch(info)
				case types.WithDrawNeoUnLockedDone:
					//	//todo wait for timeout
					e.withdrawFetch(info)
				}
			}
		}
	}
}

// withdraw fetch, from eth
func (e *EventAPI) depositFetch(info *types.LockerInfo) {
	if eth.IsConfirmedOverHeightInterval(int64(info.LockedErc20Height), ethDepositInterval, e.ethClient) {
		e.logger.Infof("deposit timeout, rHash[%s], lockerState[%s], lockerHeight[%d, %d]", info.RHash, types.LockerStateToString(types.DepositEthLockedDone), info.LockedErc20Height, ethDepositInterval)
		tx, err := eth.WrapperFetch(info.RHash, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.ethClient)
		if err != nil {
			e.logger.Errorf("depositFetch/wrapperFetch: %s", err)
			return
		}
		e.logger.Infof("deposit/wrapper fetch tx(eth): %s [%s]", tx, info.RHash)
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
func (e *EventAPI) withdrawFetch(info *types.LockerInfo) {
	if neo.IsConfirmedOverHeightInterval(info.LockedNep5Height, neoWithdrawInterval, e.neoTransaction) {
		e.logger.Infof("withdraw timeout, rHash[%s], lockerState[%s], lockerHeight[%d, %d]", info.RHash, types.LockerStateToString(types.WithDrawNeoLockedDone), info.LockedNep5Height, neoWithdrawInterval)
		tx, err := neo.WrapperFetch(info.RHash, e.cfg.NEOCfg.WIF, e.neoTransaction)
		if err != nil {
			e.logger.Errorf("loopLockerState/wrapperFetch(neo): %s", err)
			return
		}
		e.logger.Infof("withdraw/wrapper fetch tx(neo): %s [%s]", tx, info.RHash)
		info.State = types.WithDrawNeoFetchPending
		info.UnlockedNep5Hash = tx
		if err := e.store.UpdateLockerInfo(info); err != nil {
			e.logger.Error(err)
			return
		}
		e.logger.Infof("[%s] set state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchPending))

		b, height, err := neo.TxVerifyAndConfirmed(tx, neoConfirmedHeight, e.neoTransaction)
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
