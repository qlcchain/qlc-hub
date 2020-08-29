package apis

import (
	"context"
	"encoding/hex"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/qlcchain/qlc-hub/pkg/eth"
)

func (e *EventAPI) ethEventLister() {
	go func() {
		contractAddress := common.HexToAddress(e.contractAddr)
		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
		}
		filterer, err := eth.NewQLCChainFilterer(contractAddress, e.ethClient)
		if err != nil {
			e.logger.Error("NewQLCChainFilterer: ", err)
			return
		}
		logs := make(chan types.Log)
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
					e.logger.Errorf("ParseLockedState: ", err)
					break
				}
				rHash := hex.EncodeToString(event.RHash[:])
				state := event.State.Int64()
				txHash := vLog.TxHash.Hex()
				e.logger.Infof("log event: block(%s), height(%d), txHash(%s), rHash(%s)", vLog.BlockHash.Hex(), vLog.BlockNumber, vLog.TxHash.Hex(), rHash)
				processHash(state, rHash, txHash)
			}
		}
	}()
}

func processHash(state int64, rHash, txHash string) {
	switch eth.State(state) {
	case eth.IssueLock:
	case eth.IssueUnlock:
	case eth.IssueFetch:
	case eth.DestroyLock:
	case eth.DestroyUnlock:
	case eth.DestroyFetch:
	}
}
