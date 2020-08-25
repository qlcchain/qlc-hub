package wrapper

import (
	"errors"
	_ "github.com/astaxie/beego/orm"
	_ "github.com/qlcchain/qlc-hub/config"
	_ "github.com/qlcchain/qlc-hub/log"
	_ "github.com/qlcchain/qlc-hub/services/context"
	_ "go.uber.org/zap"
)

//EventInfo
type EventInfo struct {
	DId              int64      `json:"dbid"`
	Type             int64      `json:"type"`
	Status           int64      `json:"status"`
	Errno            int64      `json:"error"`
	Amount           int64      `json:"amount"`
	StartTime        int64      `json:"starttime"`
	EndTime          int64      `json:"endtime"`
	UserLockNum      int64      `json:"userlocknum"`
	WrapperLockNum   int64      `json:"wrapperlocknum"`
	LockBlockNum     int64      `json:"lockblocknum"`
	UnlockBlockNum   int64      `json:"unlockblocknum"`
	UserAccount      string     `json:"useraccount"`
	LockHash         string     `json:"lockhash"`
	HashSource       string     `json:"hashsource"`
	NeoLockTxhash    string     `json:"neolocktxhash"`
	NeoUnlockTxhash  string     `json:"neounlocktxhash"`
	NeoRefundTxhash  string     `json:"neorefundtxhash"`
	EthLockTxhash    string     `json:"ethlocktxhash"`
	EthUnlockTxhash  string     `json:"ethunlocktxhash"`
	EthDestoryTxhash string     `json:"ethdestorytxhash"`
	EventChan        chan int64 `json:"eventchan"`
}

//WrapperEventStatusChange
func (w *WrapperServer) WrapperEventStatusChange(newstatus, eventtype int64, lockhash string) (err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		w.logger.Error("WrapperEventStatusChange bad lockhash:", lockhash)
		return errors.New("bad lockhash")
	}
	if newstatus < cchEthRedemptionStatusInit || newstatus > cchEthRedemptionStatusFailed {
		w.logger.Error("WrapperEventStatusChange bad newstatus:", newstatus)
		return errors.New("bad newstatus")
	}
	if eventtype == cchEventTypeMortgage {
		if MortgageEvent[lockhash] == nil {
			w.logger.Error("WrapperEventStatusChange MortgageEvent lockhash nil", lockhash)
			return errors.New("bad MortgageEvent lockhash")
		}
		MortgageEvent[lockhash].EventChan <- newstatus
	} else if eventtype == cchEventTypeRedemption {
		if RedemptionEvent[lockhash] == nil {
			w.logger.Error("WrapperEventStatusChange RedemptionEvent lockhash nil", lockhash)
			return errors.New("bad RedemptionEvent lockhash")
		}
		RedemptionEvent[lockhash].EventChan <- newstatus
	}
	return nil
}

//WrapperEventAction  running event update
func (w *WrapperServer) WrapperEventAction(oldstatus int64, event *EventInfo) {
	if event.Type == cchEventTypeMortgage {
		switch event.Status {
		//init status, verify txhash
		case cchNep5MortgageStatusInit: //init status,unused
		case cchNep5MortgageStatusWaitNeoLockVerify: //等待neo链上lock数据确认
			txstatus, err := w.nta.Nep5TransactionVerify(event.NeoLockTxhash)
			if txstatus == CchTransactionVerifyStatusFalse {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrNep5MortgageLockFailed
				w.logger.Error("NeoLock: tx verify failed", err)
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
			_, txid, _, err := w.WrapperNep5WrapperUnlock(event.HashSource, event.UserAccount)
			if err != nil {
				w.logger.Error("WrapperNep5WrapperUnlock failed")
			} else {
				event.NeoUnlockTxhash = txid
			}
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
	} else if event.Type == cchEventTypeRedemption {
		switch event.Status {
		case cchNep5MortgageStatusInit: //unused
		case cchEthRedemptionStatusWaitEthLockVerify: //等待eth链上lock数据确认,unused
		case cchEthRedemptionStatusTryNeoLock: //准备调用neo contrack lock
			_, txid, _, err := w.WrapperNep5WrapperLock(event.Amount, event.UserLockNum, event.UserAccount, event.NeoLockTxhash)
			if err != nil {
				w.logger.Error("WrapperNep5WrapperLock failed")
			} else {
				event.NeoLockTxhash = txid
			}
		case cchEthRedemptionStatusWaitNeoLockVerify: //等待neo链上lock数据确认
			txstatus, err := w.nta.Nep5TransactionVerify(event.NeoLockTxhash)
			if txstatus == CchTransactionVerifyStatusFalse {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrNep5MortgageLockFailed
				w.logger.Error("NeoLock: tx verify failed", err)
			} else {
				event.Status = cchNep5MortgageStatusTryEthLock
			}
		case cchEthRedemptionStatusWaitClaim: //neo lock完成，等待用户claim
		case cchEthRedemptionStatusWaitNeoUnlockVerify: //等待neo链上unlock数据确认
			txstatus, err := w.nta.Nep5TransactionVerify(event.NeoLockTxhash)
			if txstatus == CchTransactionVerifyStatusFalse {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrNep5MortgageLockFailed
				w.logger.Error("NeoUnlock: tx verify failed", err)
			} else {
				event.Status = cchNep5MortgageStatusTryEthLock
			}
		case cchEthRedemptionStatusTryEthBlackhole: //准备调用eth unlock 销毁之前锁定的用户erc20 token
			_, txhash, err := w.WrapperEthDestoryUnlock(event.LockHash, event.HashSource)
			if err != nil {
				w.logger.Error("WrapperEthDestoryUnlock:err", err)
			} else {
				event.EthDestoryTxhash = txhash
			}
		case cchEthRedemptionStatusWaitEthUnlockVerify: //eth unlock数据验证 走listen
		case cchEthRedemptionStatusClaimOk: //用户正常赎回erc20资产完成
			break
		case cchEthRedemptionStatusTimeoutTryUnlock: //用户在正常时间内没有claim，wrapper尝试去neo上refund对应的nep5 token
			_, txid, _, err := w.WrapperNep5WrapperRefund(event.HashSource)
			if err != nil {
				w.logger.Error("WrapperNep5WrapperRefund failed")
			} else {
				event.NeoRefundTxhash = txid
			}
		case cchEthRedemptionStatusTimeoutUnlockVerify: //用户等待neo上refund数据确认
		case cchEthRedemptionStatusTimeoutUnlockOk: //用户超时，eth上erc20资产正常释放 unused
			txstatus, err := w.nta.Nep5TransactionVerify(event.NeoLockTxhash)
			if txstatus == CchTransactionVerifyStatusFalse {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrNep5MortgageLockFailed
				w.logger.Error("WrapperRefund: tx verify failed", err)
			} else {
				event.Status = cchNep5MortgageStatusTryEthLock
			}
		case cchEthRedemptionStatusFailed: //本次赎回失败 unused
			break
		}
	}
	return
}

//WrapperEventRunning event running
func (w *WrapperServer) WrapperEventRunning(event *EventInfo) {
	var oldstatus int64
	oldstatus = -1
	//初始化执行一次
	w.WrapperEventAction(oldstatus, event)
	oldstatus = event.Status
	//之后监听在管道，等待其他监听或者定时超时事件触发状态变更
	for {
		select {
		case newstatus := <-event.EventChan:
			if newstatus < cchEthRedemptionStatusInit || newstatus > cchEthRedemptionStatusFailed {
				w.logger.Error("WrapperEventRunning: bad newstatus")
			} else if oldstatus == newstatus {
				w.logger.Debugf("WrapperEventRunning status repeat")
			} else {
				oldstatus = event.Status
				event.Status = newstatus
				w.WrapperEventAction(oldstatus, event)
				//终结状态，退出
				if newstatus == cchEthRedemptionStatusClaimOk || newstatus == cchEthRedemptionStatusTimeoutUnlockOk || newstatus == cchEthRedemptionStatusFailed {
					if event.Type == cchEventTypeMortgage {
						gWrapperStats.RunNep5Event--
					} else {
						gWrapperStats.RunErc20Event--
					}
					gWrapperStats.RunningEvent++
					close(event.EventChan)
					w.logger.Debugf("lock(%s) event ending by status(%d)", event.LockHash, newstatus)
					return
				}
			}
		}
	}
}

//WrapperEventInit
func (w *WrapperServer) WrapperRunningEventInit() {
	var mortevents []DBNeoMortgageEventTBL
	var redemevents []DBEthRedemptionEventTBL
	var lockhash string
	var event *EventInfo

	//MortgageEvent init
	row1, err := w.sc.ocon.Raw("SELECT * from neomortgage_event_tbl Where status <?;", cchNep5MortgageStatusTimeoutDestroyOk).QueryRows(&mortevents)
	if err != nil {
		w.logger.Debugf("neomortgage_event_tbl query err")
		return
	}
	for i := 0; i < int(row1); i++ {
		lockhash = mortevents[i].LockHash
		if MortgageEvent[lockhash] != nil {
			w.logger.Debugf("WrapperEventInit mortevents repeat lockhash(%s)", lockhash)
			continue
		}
		event = new(EventInfo)
		event.Status = mortevents[i].Status
		event.Amount = mortevents[i].Amount
		event.StartTime = mortevents[i].StartTime
		event.UserLockNum = mortevents[i].UserLockNum
		event.WrapperLockNum = mortevents[i].WrapperLockNum
		event.LockBlockNum = mortevents[i].LockBlockNum
		event.UnlockBlockNum = mortevents[i].UnlockBlockNum
		event.UserAccount = mortevents[i].NeoAccount
		event.LockHash = mortevents[i].LockHash
		event.NeoLockTxhash = mortevents[i].NeoLockTxhash
		event.EthLockTxhash = mortevents[i].EthLockTxhash
		event.EventChan = make(chan int64)
		MortgageEvent[lockhash] = event
		gWrapperStats.RunNep5Event++
		gWrapperStats.RunningEvent++
	}

	//RedemptionEvent init
	row2, err := w.sc.ocon.Raw("SELECT * from neomortgage_event_tbl Where status <?;", cchEthRedemptionStatusTimeoutUnlockOk).QueryRows(&redemevents)
	if err != nil {
		w.logger.Debugf("neomortgage_event_tbl query err")
		return
	}
	for i := 0; i < int(row2); i++ {
		lockhash = redemevents[i].LockHash
		if RedemptionEvent[lockhash] != nil {
			w.logger.Debugf("WrapperEventInit redemevents repeat lockhash(%s)", lockhash)
			continue
		}
		event = new(EventInfo)
		event.Status = redemevents[i].Status
		event.Amount = redemevents[i].Amount
		event.StartTime = redemevents[i].StartTime
		event.UserLockNum = redemevents[i].UserLockNum
		event.WrapperLockNum = redemevents[i].WrapperLockNum
		event.LockBlockNum = redemevents[i].LockBlockNum
		event.UnlockBlockNum = redemevents[i].UnlockBlockNum
		event.UserAccount = redemevents[i].EthAccount
		event.LockHash = redemevents[i].LockHash
		event.NeoLockTxhash = redemevents[i].NeoLockTxhash
		event.EthLockTxhash = redemevents[i].EthLockTxhash
		event.EventChan = make(chan int64)
		RedemptionEvent[lockhash] = event
		gWrapperStats.RunErc20Event++
		gWrapperStats.RunningEvent++
	}

	return
}

func (w *WrapperServer) WrapperRunEventLimitCheck(etype int64, uaccount string) (result int) {
	count, err := w.sc.DbRunEventNumGetByUser(etype, uaccount)
	if err != nil {
		return CchRunEventCheckErr
	}
	if count >= WrapperPeruserRunEventNumLimit {
		w.logger.Debugf("user(%s) running event(%d) already over(%d)", uaccount, etype, count)
		return CchRunEventCheckFailed
	}
	return CchRunEventCheckOK
}
