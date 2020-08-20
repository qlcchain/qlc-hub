package wrapper

import (
	_ "errors"
	_ "github.com/qlcchain/qlc-hub/config"
	_ "github.com/qlcchain/qlc-hub/log"
	_ "github.com/qlcchain/qlc-hub/services/context"
	_ "go.uber.org/zap"
)

//WrapperEventRunning event running
func (w *WrapperServer) WrapperEventRunning(event *EventInfo) {
	if event.Type == cchEventTypeRedemption {
		for event.Errno != CchEventRunErrOk {
			switch event.Status {
			//init status, verify txhash
			case cchNep5MortgageStatusInit: //init status,unused
			case cchNep5MortgageStatusWaitNeoLockVerify: //等待neo链上lock数据确认
				txstatus := w.Nep5TransactionVerifyLoop(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					event.Status = cchNep5MortgageStatusFailed
					event.Errno = CchEventRunErrNep5MortgageLockFailed
					w.logger.Error("WrapperEventRunning: tx verify failed")
				} else {
					event.Status = cchNep5MortgageStatusTryEthLock
				}
			case cchNep5MortgageStatusTryEthLock: //准备调用eth contrack lock
				_, txhash, err := w.WrapperEthIssueLock(event.Amount, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEthIssueLock:err")
				} else {
					event.EthLockTxhash = txhash
				}
			case cchNep5MortgageStatusWaitEthLockVerify: //等待eth链上lock数据确认,eth listen
			case cchNep5MortgageStatusWaitClaim: //ethlock完成，等待用户claim
			case cchNep5MortgageStatusWaitEthUnlockVerify: //等待eth链上unlock数据确认,eth listen
			case cchNep5MortgageStatusTryNeoUnlock: //wrapper尝试调用neo unlock to wrapper
			case cchNep5MortgageStatusWaitNeoUnlockVerify: //等待neo链上unlock数据确认
			case cchNep5MortgageStatusClaimOk: //用户正常换取erc20资产完成
				break
			case cchNep5MortgageStatusTimeoutTryDestroy: //用户在正常时间内没有claim，wrapper尝试去eth上destroy对应的erc20资产
				_, txhash, err := w.WrapperEthIssueFetch(event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEthIssueFetch:err")
				} else {
					event.EthUnlockTxhash = txhash
				}
			case cchNep5MortgageStatusTimeoutDestroyVerify: //用户等待eth上destory数据确认,eth listen
			case cchNep5MortgageStatusTimeoutDestroyOk: //用户超时，eth上erc20资产正常销毁
				break
			case cchNep5MortgageStatusFailed: //本次抵押失败
				break
			}
		}
	} else if event.Type == cchEventTypeRedemption {
		for event.Errno != CchEventRunErrOk {
			switch event.Status {
			case cchNep5MortgageStatusInit: //unused
			case cchEthRedemptionStatusWaitEthLockVerify: //等待eth链上lock数据确认,unused
			case cchEthRedemptionStatusTryNeoLock: //准备调用neo contrack lock
				txstatus := w.Nep5ContractWrapperLock(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					w.logger.Error("Nep5ContractWrapperLock failed")
				}
			case cchEthRedemptionStatusWaitNeoLockVerify: //等待neo链上lock数据确认
				txstatus := w.Nep5TransactionVerifyLoop(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					event.Status = cchNep5MortgageStatusFailed
					event.Errno = CchEventRunErrNep5MortgageLockFailed
					w.logger.Error("WrapperEventRunning: tx verify failed")
				} else {
					event.Status = cchNep5MortgageStatusTryEthLock
				}
			case cchEthRedemptionStatusWaitClaim: //neo lock完成，等待用户claim
			case cchEthRedemptionStatusWaitNeoUnlockVerify: //等待neo链上unlock数据确认
				txstatus := w.Nep5TransactionVerifyLoop(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					event.Status = cchNep5MortgageStatusFailed
					event.Errno = CchEventRunErrNep5MortgageLockFailed
					w.logger.Error("WrapperEventRunning: tx verify failed")
				} else {
					event.Status = cchNep5MortgageStatusTryEthLock
				}
			case cchEthRedemptionStatusTryEthBlackhole: //准备调用eth unlock 销毁之前锁定的用户erc20 token
			case cchEthRedemptionStatusWaitEthUnlockVerify: //eth unlock数据验证
				txstatus := w.Nep5TransactionVerifyLoop(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					event.Status = cchNep5MortgageStatusFailed
					event.Errno = CchEventRunErrNep5MortgageLockFailed
					w.logger.Error("WrapperEventRunning: tx verify failed")
				} else {
					event.Status = cchNep5MortgageStatusTryEthLock
				}
			case cchEthRedemptionStatusClaimOk: //用户正常赎回erc20资产完成
				break
			case cchEthRedemptionStatusTimeoutTryUnlock: //用户在正常时间内没有claim，wrapper尝试去eth上unlock对应的erc20 token
			case cchEthRedemptionStatusTimeoutUnlockVerify: //用户等待eth上unlock数据确认 eth listen
			case cchEthRedemptionStatusTimeoutUnlockOk: //用户超时，eth上erc20资产正常释放 unused
			case cchEthRedemptionStatusFailed: //本次赎回失败 unused
			}
		}
	}
	return
}
