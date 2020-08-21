package wrapper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	_ "fmt"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/services/context"
	"go.uber.org/zap"
	_ "strings"
	"time"
)

var (
	MortgageEvent   map[string]*EventInfo //key -lockhash
	RedemptionEvent map[string]*EventInfo //key -lockhash
)

type WrapperServer struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
	sc     *WrapperSqlconn
	nta    *Transaction
}

var gWrapperConfig WrapperConfig

//neo contract ownner account
const WrapperNeoAccount string = "0xARJZeUehdrFD3Koy3iAymfLDWi3HtCVKYV"

//eth contract ownner account
const WrapperEthAccount string = "0x0A8EFAacbeC7763855b9A39845DDbd03b03775C1"

//eth user account
const WrapperEthUserAccount string = "0x4cD7459d7D228708C090D7d5Dc7ceDF58Cd2cD49"

//contract user prikey securt
var WrapperEthPrikey string = ""
var WrapperEthUserPrikey string = ""
var WrapperNeoPrikey string = ""

//主网合约地址： 0d821bd7b6d53f5c2b40e217c6defc8bbe896cf5
//测试网合约地址： b9d7ea3062e6aeeb3e8ad9548220c4ba1361d263
const WrapperNeoContract string = "0xb9d7ea3062e6aeeb3e8ad9548220c4ba1361d263"

//eth 测试合约地址
const WrapperEthContract string = "0xCD60c41De542ebaF81040A1F50B6eFD4B1547d91"
const wrapperEventLimit int64 = 8
const wrapperLockNum int64 = 256

const WrapperLockHashMinLen int = 32
const WrapperSourceTextMinLen int = 20
const WrapperTxHashMinLen int = 32
const WrapperAmountMinNum int = 1
const WrapperEthAddressMinNum int = 40
const WrapperLockHashHexLen int = 64

type WrapperConfig struct {
	LockNum     int64  `json:"locknum"`
	EventLimit  int64  `json:"eventlimit"`
	NeoAccount  string `json:"neoaccount"`
	NeoPrikey   string `json:"neoprikey"`
	NeoContract string `json:"neocontract"`
	EthAccount  string `json:"ethaccount"`
	EthPrikey   string `json:"ethprikey"`
	EthContract string `json:"ethcontract"`
}

type ServerInfo struct {
	TotalEvent    int64 `json:"totalnum"`
	RunningEvent  int64 `json:"runningnum"`
	RunNep5Event  int64 `json:"nep5num"`
	RunErc20Event int64 `json:"erc20num"`
}

type EventInfo struct {
	DId              int64  `json:"dbid"`
	Type             int64  `json:"type"`
	Status           int64  `json:"status"`
	Errno            int64  `json:"error"`
	Amount           int64  `json:"amount"`
	StartTime        int64  `json:"starttime"`
	EndTime          int64  `json:"endtime"`
	UserLockNum      int64  `json:"userlocknum"`
	WrapperLockNum   int64  `json:"wrapperlocknum"`
	LockBlockNum     int64  `json:"lockblocknum"`
	UnlockBlockNum   int64  `json:"unlockblocknum"`
	UserAccount      string `json:"useraccount"`
	LockHash         string `json:"lockhash"`
	HashSource       string `json:"hashsource"`
	NeoLockTxhash    string `json:"neolocktxhash"`
	NeoUnlockTxhash  string `json:"neounlocktxhash"`
	NeoRefundTxhash  string `json:"neorefundtxhash"`
	EthLockTxhash    string `json:"ethlocktxhash"`
	EthUnlockTxhash  string `json:"ethunlocktxhash"`
	EthDestoryTxhash string `json:"ethdestorytxhash"`
}

//NewWrapperServer wrapper server init
func NewWrapperServer(cfgFile string) *WrapperServer {
	cc := context.NewServiceContext(cfgFile)
	cfg, _ := cc.Config()
	wsc := NewWrapperSqlconn()
	nt := NewTransaction(neoEndPoint, WrapperNeoContract, nil)
	was := &WrapperServer{
		cfg:    cfg,
		logger: log.NewLogger("wrapper Server"),
		sc:     wsc,
		nta:    nt,
	}
	return was
}

//WrapperEventInit wrapper event init
func (w *WrapperServer) WrapperEventInit() {
	MortgageEvent = make(map[string]*EventInfo)
	RedemptionEvent = make(map[string]*EventInfo)

	//wrapperConfig initialized
	gWrapperConfig.EventLimit = wrapperEventLimit
	gWrapperConfig.LockNum = wrapperLockNum
	gWrapperConfig.NeoAccount = WrapperNeoAccount
	gWrapperConfig.NeoPrikey = WrapperNeoPrikey
	gWrapperConfig.EthAccount = WrapperEthAccount
	gWrapperConfig.EthPrikey = WrapperEthPrikey
	gWrapperConfig.NeoContract = WrapperNeoContract
	gWrapperConfig.EthContract = WrapperEthContract
	//w.logger.Debugf("get EthPrikey:%s NeoPrikey:%s EthUserPrikey:%s",gWrapperConfig.EthPrikey,gWrapperConfig.NeoPrikey,WrapperEthUserPrikey)
	go w.WrapperEthListen()
}

//WrapperSha256Get get sha256
func (w *WrapperServer) WrapperSha256Get(source string) (sum string) {
	h := sha256.New()
	h.Write([]byte(source))
	hashInBytes := h.Sum(nil)
	hashValue := hex.EncodeToString(hashInBytes)
	return hashValue
}

//WrapperEventInsert insert new event node
func (w *WrapperServer) WrapperEventInsert(stat, amount, eventType, userLocknum int64, lockHash, txHash, account string) (err error) {
	var newEvent EventInfo
	if len(txHash) < WrapperTxHashMinLen {
		return errors.New("Bad txHash")
	}
	if len(lockHash) < WrapperLockHashMinLen {
		return errors.New("Bad lockHash")
	}
	if eventType == cchEventTypeMortgage {
		if MortgageEvent[lockHash] != nil {
			return errors.New("lockHash exist")
		}
		for _, event := range MortgageEvent {
			if event.NeoLockTxhash == txHash {
				return errors.New("txHash exist")
			}
		}
	} else if eventType == cchEventTypeRedemption {
		if RedemptionEvent[lockHash] != nil {
			return errors.New("lockHash exist")
		}
		for _, event := range RedemptionEvent {
			if event.EthLockTxhash == txHash {
				return errors.New("txHash exist")
			}
		}
	} else {
		return errors.New("Bad eventType")
	}
	newEvent.Type = eventType
	newEvent.LockHash = lockHash
	newEvent.UserAccount = account
	if eventType == cchEventTypeMortgage {
		newEvent.NeoLockTxhash = txHash

	} else {
		newEvent.EthLockTxhash = txHash
	}
	newEvent.StartTime = time.Now().Unix()
	newEvent.Amount = amount
	newEvent.Status = stat
	if eventType == cchEventTypeMortgage {
		MortgageEvent[lockHash] = &newEvent
	} else if eventType == cchEventTypeRedemption {
		RedemptionEvent[lockHash] = &newEvent
	}
	nid, err := w.sc.WsqlEventRecordInsert(&newEvent)
	if err != nil {
		w.logger.Error("WsqlEventRecordInsert err", err)
	} else {
		newEvent.DId = nid
	}
	return err
}

//WrapperEventGetByLockhash get event  by lockHash
func (w *WrapperServer) WrapperEventGetByLockhash(eventType int64, lockHash string) (e *EventInfo, err error) {
	if len(lockHash) < WrapperLockHashMinLen {
		w.logger.Error("WrapperEventGetByLockhash err", lockHash)
		return nil, errors.New("Bad lockHash")
	}
	event, err := w.sc.DbGetEventByLockhash(eventType, lockHash)
	if err != nil {
		return nil, errors.New("DbGetEventByLockhash failed")
	}
	return event, nil
}

//WrapperEventGetByTxhash get event  by txHash
func (w *WrapperServer) WrapperEventGetByTxhash(eventType int64, txHash string) (event *EventInfo, err error) {
	if len(txHash) < WrapperTxHashMinLen {
		return nil, errors.New("Bad txHash")
	}
	if eventType == cchEventTypeMortgage {
		for _, event := range MortgageEvent {
			if event.NeoLockTxhash == txHash {
				return event, nil
			}
		}
	} else if eventType == cchEventTypeRedemption {
		for _, event := range RedemptionEvent {
			if event.EthLockTxhash == txHash {
				return event, nil
			}
		}
	} else {
		return nil, errors.New("Bad eventType")
	}
	return nil, errors.New("no lockHash")
}

//WrapperEventUpdateStatByTxhash update event status by tx_hash
func (w *WrapperServer) WrapperEventUpdateStatByTxhash(eventType int64, txHash string, status int64) (err error) {
	if len(txHash) < WrapperTxHashMinLen {
		return errors.New("Bad txHash")
	}
	if eventType == cchEventTypeMortgage {
		if v, ok := MortgageEvent[txHash]; ok {
			v.Status = status
			return nil
		}
	} else if eventType == cchEventTypeRedemption {
		if v, ok := RedemptionEvent[txHash]; ok {
			v.Status = status
			return nil
		}
	} else {
		return errors.New("Bad eventType")
	}
	return errors.New("no txHash")
}

//WrapperEventUpdateStatByLockhash update event status by lockHash
func (w *WrapperServer) WrapperEventUpdateStatByLockhash(eventType, status, errno int64, lockhash string) (err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		return errors.New("Bad lockhash")
	}
	if eventType != cchEventTypeMortgage && eventType != cchEventTypeRedemption {
		return errors.New("Bad eventType")
	}
	_, err = w.sc.WsqlEventDbStatusUpdate(eventType, status, errno, lockhash)
	if err != nil {
		w.logger.Error("WsqlEventDbStatusUpdate err", err)
		return errors.New("WsqlEventDbStatusUpdate err")
	}
	return w.WrapperEventUpdateCacheStatByLockhash(eventType, status, errno, lockhash)
}

//WrapperEventUpdateCacheStatByLockhash update event cache status by lockHash
func (w *WrapperServer) WrapperEventUpdateCacheStatByLockhash(eventType, status, errno int64, lockHash string) (err error) {
	if eventType == cchEventTypeMortgage {
		for _, event := range MortgageEvent {
			if event.LockHash == lockHash {
				event.Status = status
				event.Errno = errno
				return nil
			}
		}
	} else if eventType == cchEventTypeRedemption {
		for _, event := range RedemptionEvent {
			if event.LockHash == lockHash {
				event.Status = status
				event.Errno = errno
				return nil
			}
		}
	} else {
		return errors.New("Bad eventType")
	}
	return errors.New("no txHash")
}

//WrapperOnline Wrapper Online
func (w *WrapperServer) WrapperOnline() (neoaccount, neocontract, ethaccount, ethcontract string, activetime int64) {
	//w.logger.Debugf("WrapperOnline")
	return gWrapperConfig.NeoAccount, gWrapperConfig.NeoContract, gWrapperConfig.EthAccount, gWrapperConfig.EthContract, time.Now().Unix()
}

//WrapperNep5LockNotice wrapper nep5 lock notice
func (w *WrapperServer) WrapperNep5LockNotice(eventType, amount, userLocknum int64, txHash, lockHash string) (result int64) {
	if eventType != cchEventTypeMortgage && eventType != cchEventTypeRedemption {
		return cchLockNoticeRetBadParams
	}
	if int(amount) < WrapperAmountMinNum {
		return cchLockNoticeRetBadParams
	}
	err := w.WrapperEventInsert(cchNep5MortgageStatusInit, amount, eventType, userLocknum, lockHash, txHash, "")
	if err != nil {
		w.logger.Error("WrapperEventInsert err", err)
		return cchLockNoticeRetRepeat
	}
	return cchLockNoticeRetOK
}

//WrapperEthIssueLock eth IssueLock
func (w *WrapperServer) WrapperEthIssueLock(amount int64, lockhash string) (result int64, txhash string, err error) {
	if int(amount) < WrapperAmountMinNum {
		return CchEthIssueRetBadParams, "", errors.New("bad amount")
	}
	if len(lockhash) < WrapperLockHashMinLen {
		return CchEthIssueRetBadLockHash, "", errors.New("bad lockhash")
	}
	return w.EthContractIssueLock(amount, lockhash)
}

//WrapperEthIssueFetch eth IssueFetch
func (w *WrapperServer) WrapperEthIssueFetch(lockhash string) (result int64, txhash string, err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		w.logger.Error("bad lockhash")
		return CchEthIssueRetBadLockHash, "", errors.New("bad lockhash")
	}
	return w.EthContractIssueFetch(lockhash)
}

//WrapperEthDestoryUnlock eth DestoryUnlock
func (w *WrapperServer) WrapperEthDestoryUnlock(lockhash string, locksource string) (result int64, txhash string, err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		w.logger.Error("bad lockhash")
		return CchEthIssueRetBadLockHash, "", errors.New("bad lockhash")
	}
	if len(locksource) < WrapperSourceTextMinLen {
		w.logger.Error("bad locksource")
		return CchEthIssueRetBadLockHash, "", errors.New("bad locksource")
	}
	return w.EthContractDestoryUnlock(lockhash, locksource)
}

//WrapperEthUcallerDestoryLock eth user caller Destorylock
func (w *WrapperServer) WrapperEthUcallerDestoryLock(amount int64, lockhash string) (result int64, txhash string, err error) {
	if int(amount) < WrapperAmountMinNum {
		return CchEthIssueRetBadParams, "", errors.New("bad amount")
	}
	if len(lockhash) < WrapperLockHashMinLen {
		return CchEthIssueRetBadLockHash, "", errors.New("bad lockhash")
	}
	return w.EthContractUcallerDestoryLock(amount, lockhash)
}

//WrapperEthGetTransationInfo eth GetTransationInfo by tx_hash
func (w *WrapperServer) WrapperEthGetTransationInfo(infotype int64, txhash string) (result int64, info string, err error) {
	if len(txhash) != WrapperLockHashHexLen {
		w.logger.Error("bad txhash")
		return CchEthIssueRetBadTxHash, "", errors.New("no txHash")
	}
	if infotype < CchEthGetTransTypeGetAll || infotype > CchEthGetTransTypeErr {
		w.logger.Error("bad infotype:", infotype)
		return CchEthIssueRetBadParams, "", errors.New("bad infotype")
	}
	return w.EthGetBlockByTxhash(infotype, txhash)
}

//WrapperEthGetAccountInfo eth GetAccountInfo by tx_hash
func (w *WrapperServer) WrapperEthGetAccountInfo(address string) (result int64, info string, err error) {
	if len(address) < WrapperEthAddressMinNum {
		w.logger.Error("bad address")
		return CchEthIssueRetBadParams, "", errors.New("bad address")
	}
	return w.EthGetAccountByAddr(address)
}

//WrapperEthGetHashTimer eth Get HashTimer by lockhash
func (w *WrapperServer) WrapperEthGetHashTimer(lockhash string) (result, stat, amount, locknum, unlocknum int64, account, locksource string, err error) {
	if len(lockhash) < WrapperLockHashHexLen {
		w.logger.Error("bad lockhash")
		return CchEthIssueRetBadParams, 0, 0, 0, 0, "", "", errors.New("bad lockhash")
	}
	return w.EthGetHashTimer(lockhash)
}

//WrapperNep5WrapperLock neo lock token
func (w *WrapperServer) WrapperNep5WrapperLock(amount, blocknum int64, ethaddress, lockhash string) (result int64, txhash, msg string, err error) {
	var ret int64
	var transhash string
	var retmsg string
	ret = CchNeoIssueRetOK
	transhash = "0x4b92d3cabc33c4fc7d82c4c806eb77b879956018b2a9b9071da6b4ffcf4a4e25"
	retmsg = "test data"
	return ret, transhash, retmsg, nil
}

//WrapperNep5WrapperUnlock neo unlock nep5 token
func (w *WrapperServer) WrapperNep5WrapperUnlock(ethaddress, locksource string) (result int64, txhash, msg string, err error) {
	var ret int64
	var transhash string
	var retmsg string
	ret = CchNeoIssueRetOK
	transhash = "0x4b92d3cabc33c4fc7d82c4c806eb77b879956018b2a9b9071da6b4ffcf4a4e25"
	retmsg = "test data"
	return ret, transhash, retmsg, nil
}

//WrapperNep5WrapperRefund refund nep5 token
func (w *WrapperServer) WrapperNep5WrapperRefund(locksource string) (result int64, txhash, msg string, err error) {
	var ret int64
	var transhash string
	var retmsg string
	ret = CchNeoIssueRetOK
	transhash = "0x4b92d3cabc33c4fc7d82c4c806eb77b879956018b2a9b9071da6b4ffcf4a4e25"
	retmsg = "test data"
	return ret, transhash, retmsg, nil
}
