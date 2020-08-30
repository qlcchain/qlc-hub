package wrapper

import (
	"errors"
	_ "github.com/astaxie/beego/orm"
	_ "github.com/qlcchain/qlc-hub/config"
	_ "github.com/qlcchain/qlc-hub/log"
	_ "github.com/qlcchain/qlc-hub/services/context"
	_ "go.uber.org/zap"
	"strconv"
	"time"
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
	ClaimOutNum      int64      `json:"claimoutnum"`
	FetchOutNum      int64      `json:"fetchoutnum"`
	EthLockNum       int64      `json:"ethlocknum"`
	NeoLockNum       int64      `json:"neolocknum"`
	UnlockNum        int64      `json:"unlocknum"`
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

//WrapperEventAction  running event update
func (w *WrapperServer) WrapperEventAction(oldstatus int64, event *EventInfo) {
	if event.Type == cchEventTypeMortgage {
		//w.logger.Debugf("WrapperEventAction:lock(%s) status(%d->%d)",event.LockHash,oldstatus,event.Status)
		switch event.Status {
		//init status, verify txhash
		case cchNep5MortgageStatusInit: //init status,unused
		case cchNep5MortgageStatusWaitNeoLockVerify: //等待neo链上lock数据确认
			txstatus, err := w.nta.Nep5VerifyByTxid(event)
			if txstatus != CchTxVerifyStatOk {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrMortgageNep5LockFailed
				w.logger.Error("NeoLock: tx verify failed", err)
			} else {
				event.Status = cchNep5MortgageStatusTryEthLock
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchNep5MortgageStatusTryEthLock: //准备调用eth contrack lock
			_, txhash, err := w.WrapperEthIssueLock(event.Amount, event.LockHash)
			if err != nil {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrMortgageEthLockFailed
				w.logger.Error("WrapperEthIssueLock:err")
			} else {
				event.Status = cchNep5MortgageStatusWaitEthLockVerify
				event.EthLockTxhash = txhash
				w.sc.DbEventUpdate(event)
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchNep5MortgageStatusWaitEthLockVerify: //等待eth链上lock数据确认,eth listen
			//这里添加getethhashtimer的处理，是为了避免由于wrapper停掉漏掉监听日志
			w.EthVerifyByLockhash(event)
		case cchNep5MortgageStatusWaitClaim: //ethlock完成，等待用户claim
			//这里只是简单的计算claimout的blocknum
			if event.EthLockNum == 0 {
				event.ClaimOutNum = event.WrapperLockNum + gWrapperStats.CurrentEthBlocknum
			} else {
				event.ClaimOutNum = event.WrapperLockNum + event.EthLockNum
			}
			if event.NeoLockNum == 0 {
				event.FetchOutNum = event.UserLockNum + gWrapperStats.CurrentNeoBlocknum
			} else {
				event.FetchOutNum = event.UserLockNum + event.NeoLockNum
			}
			w.logger.Debugf("Mortgage event(%s) ClaimOutNum(%d) FetchOutNum(%d)", event.LockHash, event.ClaimOutNum, event.FetchOutNum)
			event.Status = cchNep5MortgageStatusWaitEthUnlockVerify
			w.sc.DbEventUpdate(event)
			go w.eventStatusUpdateMsgPush(event, event.Status)
		case cchNep5MortgageStatusWaitEthUnlockVerify: //等待eth链上unlock数据确认,eth listen
			//检测claim是否超时
			w.ClaimOutTimeCheck(event)
		case cchNep5MortgageStatusTryNeoUnlock: //wrapper尝试调用neo unlock to wrapper
			_, txid, _, err := w.WrapperNep5WrapperUnlock(event.HashSource, event.UserAccount)
			if err != nil {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrMortgageNep5UnlockFailed
				w.logger.Error("WrapperNep5WrapperUnlock failed")
			} else {
				event.NeoUnlockTxhash = txid
				//先屏蔽neounlock的验证
				//event.Status = cchNep5MortgageStatusWaitNeoUnlockVerify
				event.Status = cchNep5MortgageStatusClaimOk
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchNep5MortgageStatusWaitNeoUnlockVerify: //等待neo链上unlock数据确认
		case cchNep5MortgageStatusClaimOk: //用户正常换取erc20资产完成
			event.Errno = CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
			break
		case cchNep5MortgageStatusTimeoutTryDestroy: //用户在正常时间内没有claim，wrapper尝试去eth上destroy对应的erc20资产
			_, txhash, err := w.WrapperEthIssueFetch(event.LockHash)
			if err != nil {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrMortgageEthFetchFailed
				w.logger.Error("WrapperEthIssueFetch:err")
			} else {
				event.EthUnlockTxhash = txhash
				event.Status = cchNep5MortgageStatusTimeoutDestroyVerify
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchNep5MortgageStatusTimeoutDestroyVerify: //用户等待eth上destory数据确认,eth listen
			w.EthVerifyByLockhash(event)
		case cchNep5MortgageStatusTimeoutDestroyOk: //用户超时，eth上erc20资产正常销毁
			//检测fetch是否超时
			w.FetchOutTimeCheck(event)
			break
		case cchNep5MortgageStatusTimeoutUserFetchOk:
			event.Errno = event.Errno + CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchNep5MortgageStatusFailed: //本次抵押失败
			//检测fetch是否超时
			w.FetchOutTimeCheck(event)
			break
		case cchNep5MortgageStatusFailedFetchTimeout:
			event.Errno = event.Errno + CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
			break
		case cchNep5MortgageStatusButt: //终结状态，不处理
			break
		default:
			break
		}
	} else if event.Type == cchEventTypeRedemption {
		switch event.Status {
		case cchNep5MortgageStatusInit: //unused
		case cchEthRedemptionStatusWaitEthLockVerify: //等待eth链上lock数据确认,unused
		case cchEthRedemptionStatusTryNeoLock: //准备调用neo contrack lock
			_, txid, _, err := w.WrapperNep5WrapperLock(event.Amount, event.UserLockNum, event.UserAccount, event.LockHash)
			if err != nil {
				event.Status = cchEthRedemptionStatusFailed
				event.Errno = CchEventRunErrRedemptionNep5LockFailed
				w.logger.Error("WrapperNep5WrapperLock failed")
			} else {
				event.NeoLockTxhash = txid
				event.Status = cchEthRedemptionStatusWaitNeoLockVerify
				//w.logger.Debugf("get txid(%s)",txid)
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusWaitNeoLockVerify: //等待neo链上lock数据确认
			txstatus, err := w.nta.Nep5VerifyByTxid(event)
			if txstatus != CchTxVerifyStatOk {
				event.Status = cchEthRedemptionStatusFailed
				event.Errno = CchEventRunErrRedemptionNep5LockFailed
				w.logger.Error("NeoLock: tx verify failed", err)
			} else {
				event.Status = cchEthRedemptionStatusWaitClaim
			}
			//event.Status = cchEthRedemptionStatusWaitClaim
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusWaitClaim: //neo lock完成，等待用户claim
			//这里只是简单的计算claimout的blocknum
			if event.NeoLockNum == 0 {
				event.ClaimOutNum = event.WrapperLockNum + gWrapperStats.CurrentNeoBlocknum
			} else {
				event.ClaimOutNum = event.WrapperLockNum + event.NeoLockNum
			}
			if event.EthLockNum == 0 {
				event.FetchOutNum = event.UserLockNum + gWrapperStats.CurrentEthBlocknum
			} else {
				event.ClaimOutNum = event.UserLockNum + event.EthLockNum
			}
			event.Status = cchNep5MortgageStatusWaitEthUnlockVerify
			w.logger.Debugf("Redemption event(%s) get ClaimOutNum(%d) FetchOutNum(%d)", event.LockHash, event.ClaimOutNum, event.FetchOutNum)
			w.sc.DbEventUpdate(event)
			go w.eventStatusUpdateMsgPush(event, event.Status)
		case cchEthRedemptionStatusWaitNeoUnlockVerify: //等待neo链上unlock数据确认
			//计算claim超时
			w.ClaimOutTimeCheck(event)
		case cchEthRedemptionStatusTryEthBlackhole: //准备调用eth unlock 销毁之前锁定的用户erc20 token
			_, txhash, err := w.WrapperEthDestoryUnlock(event.LockHash, event.HashSource)
			if err != nil {
				event.Status = cchEthRedemptionStatusFailed
				event.Errno = CchEventRunErrRedemptionEthDestoryLockFailed
				w.logger.Error("WrapperEthDestoryUnlock:err(%v)", err)
			} else {
				event.EthDestoryTxhash = txhash
				event.Status = cchEthRedemptionStatusWaitEthUnlockVerify
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusWaitEthUnlockVerify: //eth unlock数据验证 走listen
			//这里添加getethhashtimer的处理，是为了避免由于wrapper停掉漏掉监听日志
			w.EthVerifyByLockhash(event)
		case cchEthRedemptionStatusClaimOk: //用户正常赎回erc20资产完成
			event.Errno = CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
			break
		case cchEthRedemptionStatusTimeoutTryUnlock: //用户在正常时间内没有claim，wrapper尝试去neo上refund对应的nep5 token
			_, txid, _, err := w.WrapperNep5WrapperRefund(event.HashSource)
			if err != nil {
				event.Status = cchEthRedemptionStatusFailed
				//event.Errno = CchEventRunErrRedemptionNep5RefundFailed
				w.logger.Error("WrapperNep5WrapperRefund failed")
			} else {
				event.NeoRefundTxhash = txid
				//event.Status = cchEthRedemptionStatusTimeoutUnlockVerify
			}
			event.Status = cchEthRedemptionStatusTimeoutUnlockVerify
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusTimeoutUnlockVerify: //用户等待neo上refund数据确认
			txstatus, err := w.nta.Nep5VerifyByTxid(event)
			if txstatus != CchTxVerifyStatOk {
				event.Status = cchNep5MortgageStatusFailed
				event.Errno = CchEventRunErrRedemptionNep5RefundFailed
				w.logger.Error("WrapperRefund: tx verify failed", err)
			} else {
				event.Status = cchEthRedemptionStatusTimeoutUnlockOk
			}
			err = w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusTimeoutUnlockOk: //用户超时，eth上erc20资产正常释放
			event.Errno = event.Errno + CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
		case cchEthRedemptionStatusFailed: //本次赎回失败 unused
			//检测fetch是否超时
			w.FetchOutTimeCheck(event)
			break
		case cchEthRedemptionStatusFailedFetchTimeout: //
			event.Errno = event.Errno + CchEventRunErrEndingOk
			err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err(%v)", err)
			}
			break
		case cchEthRedemptionStatusButt:
			break
		default:
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
	//之后监听在管道，等待其他监听或者定时超时事件触发状态变更
	for {
		select {
		case newstatus := <-event.EventChan:
			w.logger.Debugf("WrapperEventRunning:event(%s) get newstatus(%d) oldstatus(%d)", event.LockHash, newstatus, oldstatus)
			if newstatus < cchEthRedemptionStatusInit || newstatus >= cchEthRedemptionStatusButt {
				w.logger.Error("WrapperEventRunning: bad newstatus")
			} else if oldstatus == newstatus {
				w.logger.Debugf("WrapperEventRunning status repeat")
			} else {
				oldstatus = event.Status
				event.Status = newstatus
				w.WrapperEventAction(oldstatus, event)
				//终结状态，退出
				if event.Type == cchEventTypeMortgage {
					if event.Errno >= CchEventRunErrEndingOk {
						gWrapperStats.RunNep5Event--
						gWrapperStats.RunningEvent--
						close(event.EventChan)
						w.logger.Debugf("Mortgage event lock(%s) event ending by status(%d) errno(%d)", event.LockHash, event.Status, event.Errno)
						return
					}
				} else {
					if event.Errno >= CchEventRunErrEndingOk {
						gWrapperStats.RunErc20Event--
						gWrapperStats.RunningEvent--
						close(event.EventChan)
						w.logger.Debugf("Redemption event lock(%s) event ending by status(%d) errno(%d)", event.LockHash, event.Status, event.Errno)
						return
					}
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
	var sqlcmd string

	//MortgageEvent init
	sqlcmd = "SELECT * from neomortgage_event_tbl Where error<" + strconv.FormatInt(CchEventRunErrEndingOk, 10) + ";"
	row1, err := w.sc.ocon.Raw(sqlcmd).QueryRows(&mortevents)
	if err != nil {
		w.logger.Debugf("neomortgage_event_tbl query err")
		return
	}
	w.logger.Debugf("sqlcmd(%s) ret(%d)", sqlcmd, row1)
	for i := 0; i < int(row1); i++ {
		lockhash = mortevents[i].LockHash
		if MortgageEvent[lockhash] != nil {
			w.logger.Debugf("WrapperEventInit mortevents repeat lockhash(%s)", lockhash)
			continue
		}
		event = new(EventInfo)
		event.Type = cchEventTypeMortgage
		event.Status = mortevents[i].Status
		event.Amount = mortevents[i].Amount
		event.StartTime = mortevents[i].StartTime
		event.UserLockNum = mortevents[i].UserLockNum
		event.WrapperLockNum = mortevents[i].WrapperLockNum
		event.EthLockNum = mortevents[i].EthLockNum
		event.NeoLockNum = mortevents[i].NeoLockNum
		event.UnlockNum = mortevents[i].UnlockNum
		event.ClaimOutNum = mortevents[i].ClaimOutNum
		event.FetchOutNum = mortevents[i].FetchOutNum
		event.UserAccount = mortevents[i].NeoAccount
		event.LockHash = mortevents[i].LockHash
		event.NeoLockTxhash = mortevents[i].NeoLockTxhash
		event.EthLockTxhash = mortevents[i].EthLockTxhash
		event.EventChan = make(chan int64)
		MortgageEvent[lockhash] = event
		gWrapperStats.RunNep5Event++
		gWrapperStats.RunningEvent++
		//w.logger.Debugf("WrapperEventRunning:mortevent(%d:%s)",event.Status,event.LockHash)
		go w.WrapperEventRunning(event)
	}

	//RedemptionEvent init
	sqlcmd = "SELECT * from ethredemption_event_tbl Where error<" + strconv.FormatInt(CchEventRunErrEndingOk, 10) + ";"
	row2, err := w.sc.ocon.Raw(sqlcmd).QueryRows(&redemevents)
	if err != nil {
		w.logger.Debugf("ethredemption_event_tbl query err")
		return
	}
	w.logger.Debugf("sqlcmd(%s) ret(%d)", sqlcmd, row2)
	for i := 0; i < int(row2); i++ {
		lockhash = redemevents[i].LockHash
		if RedemptionEvent[lockhash] != nil {
			w.logger.Debugf("WrapperEventInit redemevents repeat lockhash(%s)", lockhash)
			continue
		}
		event = new(EventInfo)
		event.Type = cchEventTypeRedemption
		event.Status = redemevents[i].Status
		event.Amount = redemevents[i].Amount
		event.StartTime = redemevents[i].StartTime
		event.UserLockNum = redemevents[i].UserLockNum
		event.WrapperLockNum = redemevents[i].WrapperLockNum
		event.EthLockNum = redemevents[i].EthLockNum
		event.NeoLockNum = redemevents[i].NeoLockNum
		event.UnlockNum = redemevents[i].UnlockNum
		event.ClaimOutNum = redemevents[i].ClaimOutNum
		event.FetchOutNum = redemevents[i].FetchOutNum
		event.UserAccount = redemevents[i].EthAccount
		event.LockHash = redemevents[i].LockHash
		event.NeoLockTxhash = redemevents[i].NeoLockTxhash
		event.EthLockTxhash = redemevents[i].EthLockTxhash
		event.EventChan = make(chan int64)
		RedemptionEvent[lockhash] = event
		gWrapperStats.RunErc20Event++
		gWrapperStats.RunningEvent++
		//w.logger.Debugf("WrapperEventRunning:redemevent(%d:%s)",event.Status,event.LockHash)
		go w.WrapperEventRunning(event)
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

//ClaimOutTimeCheck claim time out check
func (w *WrapperServer) ClaimOutTimeCheck(event *EventInfo) {
	d := time.Duration(time.Second * 10)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		<-t.C
		if event.Status == cchNep5MortgageStatusWaitEthUnlockVerify && event.Type == cchEventTypeMortgage {
			//w.logger.Debugf("get CurrentEthBlocknum(%d) ClaimOutNum(%d)",gWrapperStats.CurrentEthBlocknum,event.ClaimOutNum)
			if gWrapperStats.CurrentEthBlocknum > event.ClaimOutNum {
				w.logger.Debugf("Mortgage event hash(%s) claim timeout in block(%d)", event.LockHash, gWrapperStats.CurrentEthBlocknum)
				event.Status = cchNep5MortgageStatusTimeoutTryDestroy
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else if event.Status == cchEthRedemptionStatusWaitNeoUnlockVerify && event.Type == cchEventTypeRedemption {
			if gWrapperStats.CurrentNeoBlocknum > event.ClaimOutNum {
				w.logger.Debugf("Redemption event hash(%s) claim timeout in block(%d)", event.LockHash, gWrapperStats.CurrentNeoBlocknum)
				event.Status = cchEthRedemptionStatusTimeoutTryUnlock
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else {
			//其他状态直接退出
			return
		}
	}
}

//FetchOutTimeCheck fetch time out
func (w *WrapperServer) FetchOutTimeCheck(event *EventInfo) {
	d := time.Duration(time.Second * 10)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		<-t.C
		//w.logger.Debugf("FetchOutTimeCheck status(%d) FetchOutNum(%d) CurrentNeoBlocknum(%d)",event.Status,event.FetchOutNum,gWrapperStats.CurrentNeoBlocknum)
		if event.Status == cchNep5MortgageStatusTimeoutDestroyOk {
			if gWrapperStats.CurrentNeoBlocknum > event.FetchOutNum {
				w.logger.Debugf("Mortgage event hash(%s) fetch timeout in block(%d)", event.LockHash, gWrapperStats.CurrentNeoBlocknum)
				event.Status = cchNep5MortgageStatusTimeoutUserFetchOk
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else if event.Status == cchEthRedemptionStatusClaimOk {
			if gWrapperStats.CurrentEthBlocknum > event.FetchOutNum {
				w.logger.Debugf("Redemption event hash(%s) fetch timeout in block(%d)", event.LockHash, gWrapperStats.CurrentEthBlocknum)
				event.Status = cchEthRedemptionStatusTimeoutUserFetchOk
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else if event.Status == cchNep5MortgageStatusFailed && event.Type == cchEventTypeMortgage {
			if gWrapperStats.CurrentNeoBlocknum > event.FetchOutNum {
				w.logger.Debugf("Mortgage failedevent hash(%s) fetch timeout in block(%d)", event.LockHash, gWrapperStats.CurrentNeoBlocknum)
				event.Status = cchNep5MortgageStatusFailedFetchTimeout
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else if event.Status == cchEthRedemptionStatusFailed && event.Type == cchEventTypeRedemption {
			if gWrapperStats.CurrentEthBlocknum > event.FetchOutNum {
				w.logger.Debugf("Redemption failedevent hash(%s) fetch timeout in block(%d)", event.LockHash, gWrapperStats.CurrentEthBlocknum)
				event.Status = cchEthRedemptionStatusFailedFetchTimeout
				err := w.WrapperEventUpdateStatAndErrnoByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
				if err != nil {
					w.logger.Error("WrapperEventUpdateStatAndErrnoByLockhash: err", err)
				}
				return
			}
		} else {
			//其他状态直接退出
			return
		}

	}
}

func (w *WrapperServer) eventStatusUpdateMsgPush(event *EventInfo, newstatus int64) (err error) {
	if event == nil {
		return errors.New("bad event")
	}
	if event.Errno < CchEventRunErrEndingOk {
		event.EventChan <- newstatus
		w.logger.Debugf("event(%s) change status(%d)", event.LockHash, newstatus)
	}
	return nil
}
