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
	go func() {
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
				e.logger.Infof("log event: rHash[%s], state[%d], txHash[%s]", rHash, state, vLog.TxHash.Hex())
				go e.processEthEvent(state, rHash, txHash)
			}
		}
	}()
}

func (e *EventAPI) processEthEvent(state int64, rHash, txHash string) {
	b, _, err := eth.TxVerifyAndConfirmed(txHash, ethConfirmedHeight, e.ethClient)
	if !b || err != nil {
		e.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, rHash)
		return
	}

	switch eth.State(state) {
	case eth.IssueLock:
		info, err := e.store.GetLockerInfo(rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}

		// update indo
		hashTimer, err := eth.GetHashTimer(e.ethClient, e.ethContract, rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		info.State = types.DepositEthLockedDone
		info.LockedErc20Hash = txHash
		info.LockedErc20Height = hashTimer.LockedHeight
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthLockedDone))

		//todo notify

	case eth.IssueUnlock:
		info, err := e.store.GetLockerInfo(rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}

		// update info
		hashTimer, err := eth.GetHashTimer(e.ethClient, e.ethContract, rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		info.State = types.DepositEthUnLockedDone
		info.ROrigin = hashTimer.ROrigin
		info.UnlockedErc20Hash = txHash
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthUnLockedDone))

		//todo notify

		// to neo unlock
		tx, err := neo.WrapperUnlock(hashTimer.ROrigin, e.cfg.NEOCfg.WIF, hashTimer.UserAddr, e.neoTransaction)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		e.logger.Infof("deposit/wrapper neo unlock: %s [%s]", tx, rHash)
		info.State = types.DepositNeoUnLockedPending
		info.UnlockedNep5Hash = tx
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoUnLockedPending))

		b, height, err := neo.TxVerifyAndConfirmed(tx, neoConfirmedHeight, e.neoTransaction)
		if !b || err != nil {
			e.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, rHash)
			return
		}
		info.State = types.DepositNeoUnLockedDone
		info.UnlockedNep5Height = height
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositNeoUnLockedDone))

	case eth.IssueFetch:
		info, err := e.store.GetLockerInfo(rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}

		// update info
		hashTimer, err := eth.GetHashTimer(e.ethClient, e.ethContract, rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		info.State = types.DepositEthFetchDone
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.DepositEthFetchDone))

	case eth.DestroyLock:
		b, _, err := eth.TxVerifyAndConfirmed(txHash, ethConfirmedHeight, e.ethClient)
		if !b || err != nil {
			e.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, rHash)
			return
		}

		// add info
		hashTimer, err := eth.GetHashTimer(e.ethClient, e.ethContract, rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		fmt.Println("====== ", hashTimer.String())

		info := new(types.LockerInfo)
		info.State = types.WithDrawEthLockedDone
		info.RHash = rHash
		info.LockedErc20Hash = txHash
		info.LockedErc20Height = hashTimer.LockedHeight
		info.Amount = hashTimer.Amount.Int64()
		info.Erc20Addr = hashTimer.UserAddr
		if err := e.store.AddLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("add [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthLockedDone))

		// neo lock
		tx, err := neo.WrapperLock(e.cfg.NEOCfg.WIF, hashTimer.UserAddr, rHash, int(info.Amount), e.neoTransaction)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}
		e.logger.Info("withdraw/wrapper neo lock tx: ", tx)
		info.State = types.WithDrawNeoLockedPending
		info.LockedNep5Hash = tx
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoLockedPending))

		b, height, err := neo.TxVerifyAndConfirmed(tx, neoConfirmedHeight, e.neoTransaction)
		if !b || err != nil {
			e.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, rHash)
			return
		}
		info.State = types.WithDrawNeoLockedDone
		info.LockedNep5Height = height
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoLockedDone))
		//todo notify

	case eth.DestroyUnlock:
		info, err := e.store.GetLockerInfo(rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}

		hashTimer, err := eth.GetHashTimer(e.ethClient, e.cfg.EthereumCfg.Contract, rHash)
		if err != nil {
			e.logger.Errorf("get timer: %s [%s]", err, rHash)
			return
		}
		info.UnlockedErc20Height = hashTimer.UnlockedHeight
		info.State = types.WithDrawEthUnlockDone
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawEthUnlockDone))

	case eth.DestroyFetch:
		info, err := e.store.GetLockerInfo(rHash)
		if err != nil {
			e.logger.Errorf("processEthEvent: %s [%s]", err, rHash)
			return
		}

		tx, err := neo.RefundWrapper(info.RHash, e.cfg.NEOCfg.WIF, e.neoTransaction)
		if err != nil {
			e.logger.Errorf("processEthEvent/RefundWrapper: %s [%s]", err, rHash)
			return
		}
		e.logger.Infof("withdraw/wrapper neo fetch tx: %s [%s]", tx, info.RHash)
		info.UnlockedNep5Hash = tx
		info.State = types.WithDrawNeoFetchPending
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchPending))

		b, height, err := neo.TxVerifyAndConfirmed(tx, neoConfirmedHeight, e.neoTransaction)
		if !b || err != nil {
			e.logger.Errorf("processEthEvent: %s, %v [%s]", err, b, rHash)
			return
		}
		info.State = types.WithDrawNeoFetchDone
		info.UnlockedNep5Height = height
		if err := e.store.UpdateLockerInfo(info); err != nil {
			return
		}
		e.logger.Infof("set [%s] state to [%s]", info.RHash, types.LockerStateToString(types.WithDrawNeoFetchDone))
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
				e.logger.Errorf("loopLockerState/GetLockerInfos: %s", err)
			}
			for _, info := range infos {
				switch info.State {
				case types.DepositEthLockedDone:
					if info.LockedErc20Height-eth.GetBestBlockHeight() > uint32(ethDepositInterval) {
						tx, err := eth.WrapperFetch(info.RHash, e.cfg.EthereumCfg.Account, e.cfg.EthereumCfg.Contract, e.ethClient)
						if err != nil {
							e.logger.Errorf("loopLockerState/WrapperFetch: %s", err)
						}
						e.logger.Info("deposit/wrapper eth fetch tx: ", tx)
						info.UnlockedErc20Hash = tx
						info.State = types.DepositEthFetchPending
						if err := e.store.UpdateLockerInfo(info); err != nil {
							e.logger.Error(err)
						}
					}
					//case types.WithDrawNeoLockedDone:

				}
			}
		}
	}
}
