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
			case cchNep5MortgageStatusInit:
				txstatus := w.Nep5TransactionVerifyLoop(event.NeoLockTxhash)
				if txstatus == CchTransactionVerifyStatusFalse {
					event.Status = cchNep5MortgageStatusFailed
					event.Errno = CchEventRunErrNep5MortgageLockFailed
					w.logger.Error("WrapperEventRunning: tx verify failed")
				} else {
					event.Status = cchNep5MortgageStatusTryEthLock
				}
			}
		}
	} else if event.Type == cchEventTypeRedemption {
		for event.Errno != CchEventRunErrOk {
			switch event.Status {
			case cchNep5MortgageStatusInit:

			}
		}
	}
	return
}
