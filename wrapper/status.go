package wrapper

import (
	_ "fmt"
)

const (
	cchNep5MortgageStatusInit                 int64 = 0  //初始化状态
	cchNep5MortgageStatusWaitNeoLockVerify          = 1  //等待neo链上lock数据确认
	cchNep5MortgageStatusTryEthLock                 = 2  //准备调用eth contrack lock
	cchNep5MortgageStatusWaitEthLockVerify          = 3  //等待eth链上lock数据确认
	cchNep5MortgageStatusWaitClaim                  = 4  //ethlock完成，等待用户claim
	cchNep5MortgageStatusWaitEthUnlockVerify        = 5  //等待eth链上unlock数据确认
	cchNep5MortgageStatusTryNeoUnlock               = 6  //wrapper尝试调用neo unlock to wrapper
	cchNep5MortgageStatusWaitNeoUnlockVerify        = 7  //等待neo链上unlock数据确认
	cchNep5MortgageStatusClaimOk                    = 8  //用户正常换取erc20资产完成
	cchNep5MortgageStatusTimeoutTryDestroy          = 9  //用户在正常时间内没有claim，wrapper尝试去eth上destroy对应的erc20资产
	cchNep5MortgageStatusTimeoutDestroyVerify       = 10 //用户等待eth上destory数据确认
	cchNep5MortgageStatusTimeoutDestroyOk           = 11 //用户超时，eth上erc20资产正常销毁
	cchNep5MortgageStatusFailed                     = 12 //本次抵押失败
)

const (
	cchNep5MgErrNeoLockVerifyFailed int64 = 1 //neo链上lock数据确认失败
)

const (
	cchEthRedemptionStatusInit                int64 = 0  //初始化状态
	cchEthRedemptionStatusWaitEthLockVerify         = 1  //等待eth链上lock数据确认
	cchEthRedemptionStatusTryNeoLock                = 2  //准备调用neo contrack lock
	cchEthRedemptionStatusWaitNeoLockVerify         = 3  //等待neo链上lock数据确认
	cchEthRedemptionStatusWaitClaim                 = 4  //neo lock完成，等待用户claim
	cchEthRedemptionStatusWaitNeoUnlockVerify       = 5  //等待neo链上unlock数据确认
	cchEthRedemptionStatusTryEthBlackhole           = 6  //准备调用eth unlock 销毁之前锁定的用户erc20 token
	cchEthRedemptionStatusWaitEthUnlockVerify       = 7  //eth unlock数据验证
	cchEthRedemptionStatusClaimOk                   = 8  //用户正常赎回erc20资产完成
	cchEthRedemptionStatusTimeoutTryUnlock          = 9  //用户在正常时间内没有claim，wrapper尝试去eth上unlock对应的erc20 token
	cchEthRedemptionStatusTimeoutUnlockVerify       = 10 //用户等待eth上unlock数据确认
	cchEthRedemptionStatusTimeoutUnlockOk           = 11 //用户超时，eth上erc20资产正常释放
	cchEthRedemptionStatusFailed                    = 12 //本次赎回失败
)

const (
	cchEventTypeMortgage   int64 = 1 //抵押事件
	cchEventTypeRedemption       = 2 //赎回事件
)

const (
	cchLockNoticeRetOK        int64 = 0 //正常
	cchLockNoticeRetBadParams       = 1 //参数错误
	cchLockNoticeRetRepeat          = 2 //事件重复
)
const (
	CchGetEventStatusRetOK       int64 = 0 //正常
	CchGetEventStatusRetNoTxhash       = 1 //txhash 没找到
)

const (
	CchEthGetTransTypeGetAll int64 = 1 //
	CchEthGetTransTypeErr
)
const (
	CchEthIssueRetOK               int64 = 1 //正常
	CchEthIssueRetBadParams              = 2 //参数错误
	CchEthIssueRetBadTxHash              = 3 //TxHash 错误
	CchEthIssueRetBadLockHash            = 4 //LockHash 错误
	CchEthIssueRetClientConnFailed       = 5 // client连接失败
	CchEthIssueRetBadKey                 = 6 //bad key
)
const (
	CchNeoIssueRetOK               int64 = 1 //正常
	CchNeoIssueRetBadParams              = 2 //参数错误
	CchNeoIssueRetBadTxHash              = 3 //TxHash 错误
	CchNeoIssueRetClientConnFailed       = 4 // client连接失败
)

const (
	CchEventRunErrOk                     int64 = 0 //正常
	CchEventRunErrNep5MortgageLockFailed           //nep5 moregage trans verify failed
	CchEventRunErrnoUnknown
)

const (
	CchTransactionVerifyStatusUnknown = -1
	CchTransactionVerifyStatusFalse   = 0
	CchTransactionVerifyStatusTrue    = 1
)
const WrapperTimestampFormat = "02/01/2006 15:04:05 PM"
