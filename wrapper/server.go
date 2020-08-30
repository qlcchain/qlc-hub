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
var gWrapperStats WrapperStatistics

//neo contract ownner account
const WrapperNeoAccount string = "ARJZeUehdrFD3Koy3iAymfLDWi3HtCVKYV"

//eth contract ownner account
const WrapperEthAccount string = "0x0A8EFAacbeC7763855b9A39845DDbd03b03775C1"

//eth user account
const WrapperEthUserAccount string = "4cD7459d7D228708C090D7d5Dc7ceDF58Cd2cD49"

//contract user prikey securt
var WrapperEthPrikey string = ""
var WrapperEthUserPrikey string = ""
var WrapperNeoPrikey string = ""

//neo 合约测试地址
const WrapperNeoContract string = "c59bd98299324d6156c67cd9e2e9783054eaf383"

//eth 测试合约地址
const WrapperEthContract string = "0x6d37597F0d9e917baeF2727ece52AEeb8B5294c7"

//const WrapperEthContract string = "0xCD60c41De542ebaF81040A1F50B6eFD4B1547d91"
const wrapperEventLimit int64 = 8
const wrapperLockNum int64 = 256

const WrapperLockHashMinLen int = 32
const WrapperSourceTextMinLen int = 20
const WrapperTxHashMinLen int = 64
const WrapperHashHexStringWithPrix int = 66
const WrapperAmountMinNum int = 1
const WrapperEthAddressHexLen int = 40
const WrapperLockHashHexLen int = 64
const WrapperGasWeiNum int64 = 100000000
const WrapperPeruserRunEventNumLimit int = 5
const DefaultEthWrapperLockNum int64 = 15 //15 eth block   2min
const DefaultNeoWrapperLockNum int64 = 10 //10  neo block  150 sec
const DefaultEthUserLockNum int64 = 50    //450 sec
const DefaultNeoUserLockNum int64 = 20    //5 min

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

type WrapperStatistics struct {
	//TotalEvent    	int64 `json:"totalnum"`
	RunningEvent       int64 `json:"runningnum"`
	RunNep5Event       int64 `json:"nep5num"`
	RunErc20Event      int64 `json:"erc20num"`
	LastEthBlocknum    int64 `json:"lastethbkn"`
	LastNeoBlocknum    int64 `json:"lastneobkn"`
	CurrentEthBlocknum int64 `json:"curethbkn"`
	CurrentNeoBlocknum int64 `json:"curneobkn"`
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
	w.NeoBlockNumberSysn()
	w.EthBlockNumbersysn()
	w.WrapperRunningEventInit()
	go w.NeoUpdateBlockNumber()
	go w.EthUpdateBlockNumber()
	go w.WrapperEthListen()
	go w.nta.NeoGetNep5EventListen()
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
	newEvent := new(EventInfo)
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
		newEvent.UserLockNum = DefaultNeoUserLockNum
		newEvent.WrapperLockNum = DefaultEthWrapperLockNum
	} else {
		newEvent.EthLockTxhash = txHash
		newEvent.UserLockNum = DefaultEthUserLockNum
		newEvent.WrapperLockNum = DefaultNeoWrapperLockNum
	}
	newEvent.StartTime = time.Now().Unix()
	newEvent.Amount = amount
	newEvent.Status = stat
	newEvent.EventChan = make(chan int64)
	if eventType == cchEventTypeMortgage {
		MortgageEvent[lockHash] = newEvent
	} else if eventType == cchEventTypeRedemption {
		RedemptionEvent[lockHash] = newEvent
	}
	nid, err := w.sc.DbEventRecordInsert(newEvent)
	if err != nil {
		w.logger.Error("DbEventRecordInsert err", err)
	} else {
		newEvent.DId = nid
	}
	//init new running event
	go w.WrapperEventRunning(newEvent)
	return err
}

//WrapperEventGetByLockhash get event  by lockHash
func (w *WrapperServer) WrapperEventGetByLockhash(eventType int64, lockHash string) (e *EventInfo, err error) {
	lh, err := w.WrapperHashHexStrDelPrix(lockHash)
	if err != nil {
		w.logger.Error("WrapperEventGetByLockhash err", lockHash)
		return nil, errors.New("Bad lockHash")
	}
	event, err := w.sc.DbGetEventByLockhash(eventType, lh)
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

//WrapperEventUpdateStatAndErrnoByLockhash update event status by lockHash
func (w *WrapperServer) WrapperEventUpdateStatAndErrnoByLockhash(eventType, status, errno int64, lockhash string) (err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		return errors.New("Bad lockhash")
	}
	if eventType != cchEventTypeMortgage && eventType != cchEventTypeRedemption {
		return errors.New("Bad eventType")
	}
	_, err = w.sc.WsqlEventDbStatusAndErrnoUpdate(eventType, status, errno, lockhash)
	if err != nil {
		w.logger.Error("WsqlEventDbStatusAndErrnoUpdate err", err)
		return errors.New("WsqlEventDbStatusAndErrnoUpdate err")
	}
	return w.WrapperEventUpdateCacheStatByLockhash(eventType, status, errno, lockhash)
}

//WrapperEventUpdateCacheStatByLockhash update event cache status by lockHash
func (w *WrapperServer) WrapperEventUpdateCacheStatByLockhash(eventtype, status, errno int64, lockhash string) (err error) {
	if len(lockhash) < WrapperLockHashMinLen {
		w.logger.Error("WrapperEventUpdateCacheStatByLockhash bad lockhash:", lockhash)
		return errors.New("bad lockhash")
	}
	if status < cchEthRedemptionStatusInit || status >= cchNep5MortgageStatusButt {
		w.logger.Error("WrapperEventUpdateCacheStatByLockhash bad newstatus(%d)", status)
		return errors.New("bad newstatus")
	}
	if eventtype == cchEventTypeMortgage {
		if MortgageEvent[lockhash] == nil {
			w.logger.Error("WrapperEventUpdateCacheStatByLockhash MortgageEvent lockhash nil", lockhash)
			return errors.New("bad MortgageEvent lockhash")
		}
		//w.logger.Debugf("WrapperEventUpdateCacheStatByLockhash:(%d->%d)",MortgageEvent[lockhash].Status,status)
		if MortgageEvent[lockhash].Errno != errno {
			MortgageEvent[lockhash].Errno = errno
		}
		if MortgageEvent[lockhash].Status != status {
			MortgageEvent[lockhash].Status = status
		}
		go w.eventStatusUpdateMsgPush(MortgageEvent[lockhash], status)
		return nil
	} else if eventtype == cchEventTypeRedemption {
		if RedemptionEvent[lockhash] == nil {
			w.logger.Error("WrapperEventUpdateCacheStatByLockhash RedemptionEvent lockhash nil", lockhash)
			return errors.New("bad RedemptionEvent lockhash")
		}
		//w.logger.Debugf("WrapperEventUpdateCacheStatByLockhash:(%d->%d)",RedemptionEvent[lockhash].Status,status)
		if RedemptionEvent[lockhash].Errno != errno {
			RedemptionEvent[lockhash].Errno = errno
		}
		if RedemptionEvent[lockhash].Status != status {
			RedemptionEvent[lockhash].Status = status
		}
		go w.eventStatusUpdateMsgPush(RedemptionEvent[lockhash], status)
		return nil
	}
	return errors.New("bad eventtype")
}

//WrapperOnline Wrapper Online
func (w *WrapperServer) WrapperOnline() (neoaccount, neocontract, ethaccount, ethcontract string, activetime int64) {
	//w.logger.Debugf("WrapperOnline")
	return gWrapperConfig.NeoAccount, gWrapperConfig.NeoContract, gWrapperConfig.EthAccount, gWrapperConfig.EthContract, time.Now().Unix()
}

//WrapperNep5LockNotice wrapper nep5 lock notice
func (w *WrapperServer) WrapperNep5LockNotice(action, amount, userlocknum int64, txhash, lockhash, source string) (result int64) {
	var estatus int64
	var etype int64
	if int(amount) < WrapperAmountMinNum {
		return cchLockNoticeRetBadParams
	}
	lh, err := w.WrapperHashHexStrDelPrix(lockhash)
	if err != nil {
		return cchLockNoticeRetBadParams
	}
	switch action {
	case Nep5ActionUserLock:
		etype = cchEventTypeMortgage
		estatus = cchNep5MortgageStatusWaitNeoLockVerify
		err := w.WrapperEventInsert(estatus, amount, etype, userlocknum, lh, txhash, "")
		if err != nil {
			w.logger.Error("WrapperNep5LockNotice WrapperEventInsert err", err)
			return cchLockNoticeRetRepeat
		}
	case Nep5ActionRefundUser:
		etype = cchEventTypeMortgage
		estatus = cchNep5MortgageStatusUserFetched
		_, err := w.sc.WsqlEventDbStatusUpdate(etype, estatus, lh)
		if err != nil {
			w.logger.Error("WrapperNep5LockNotice WsqlEventDbStatusUpdate err", err)
			return cchLockNoticeRetRepeat
		}
	case Nep5ActionUserUnlock:
		lh, err := w.WrapperHashHexStrDelPrix(lockhash)
		if err != nil {
			w.logger.Errorf("WrapperNep5LockNotice bad lockhash(%s)", lockhash)
			return cchLockNoticeRetRepeat
		}
		event := RedemptionEvent[lh]
		if event == nil {
			w.logger.Errorf("WrapperNep5LockNotice lockhash found err(%s)", lh)
			return cchLockNoticeRetRepeat
		}
		event.NeoUnlockTxhash = txhash
		event.HashSource = source
		txstatus, err := w.nta.Nep5VerifyByTxid(event)
		if txstatus != CchTxVerifyStatOk {
			event.Status = cchEthRedemptionStatusFailed
			event.Errno = CchEventRunErrRedemptionNep5LockFailed
			w.logger.Error("NeoLock: tx verify failed", err)
		} else {
			event.Status = cchEthRedemptionStatusTryEthBlackhole
		}
		_, err = w.sc.DbEventUpdate(event)
		if err != nil {
			w.logger.Error("DbEventUpdate fail", err)
			return cchLockNoticeRetRepeat
		}
		go w.eventStatusUpdateMsgPush(event, event.Status)
		w.logger.Debugf("Nep5ActionUserUnlock:hash(%s) source(%s) txhash(%s) status(%d)", lh, event.HashSource, txhash, event.Status)
	default:
		w.logger.Error("WrapperNep5LockNotice:bad action(%d)", action)
		return cchLockNoticeRetBadParams
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
	if len(address) < WrapperEthAddressHexLen {
		w.logger.Error("bad address")
		return CchEthIssueRetBadParams, "", errors.New("bad address")
	}
	return w.EthGetAccountByAddr(address)
}

//WrapperEthGetHashTimer eth Get HashTimer by lockhash
func (w *WrapperServer) WrapperEthGetHashTimer(lockhash string) (result int64, elog *etheventhtimerlog, err error) {
	if len(lockhash) < WrapperLockHashHexLen {
		w.logger.Error("bad lockhash")
		return CchEthIssueRetBadParams, nil, errors.New("bad lockhash")
	}
	return w.EthGetHashTimer(lockhash)
}

//WrapperNep5WrapperLock neo lock token
func (w *WrapperServer) WrapperNep5WrapperLock(amount, blocknum int64, ethaddress, lockhash string) (result int64, txhash, msg string, err error) {
	if len(lockhash) < WrapperLockHashHexLen {
		w.logger.Error("bad lockhash %s", lockhash)
		return CchNeoIssueRetBadParams, "", "", err
	}
	txid, err := w.nta.Nep5ContractWrapperLock(amount, blocknum, ethaddress, lockhash)
	if err != nil {
		return CchNeoIssueRetBadParams, "", "", err
	}
	return CchNeoIssueRetOK, txid, "", nil
}

//WrapperNep5WrapperUnlock neo unlock nep5 token
func (w *WrapperServer) WrapperNep5WrapperUnlock(locksource, ethaddress string) (result int64, txhash, msg string, err error) {
	if len(locksource) < WrapperSourceTextMinLen {
		w.logger.Error("WrapperNep5WrapperUnlock :bad locksource")
		return CchNeoIssueRetBadParams, "", "", err
	}
	txid, err := w.nta.Nep5ContractWrapperUnlock(locksource, ethaddress)
	if err != nil {
		w.logger.Error("WrapperNep5WrapperUnlock err", err)
		return CchNeoIssueRetBadParams, "", "", err
	}
	return CchNeoIssueRetOK, txid, "", nil
}

//WrapperNep5WrapperRefund refund nep5 token
func (w *WrapperServer) WrapperNep5WrapperRefund(locksource string) (result int64, txhash, msg string, err error) {
	if len(locksource) < WrapperSourceTextMinLen {
		w.logger.Error("WrapperNep5WrapperRefund :bad locksource")
		return CchNeoIssueRetBadParams, "", "", err
	}
	txid, err := w.nta.Nep5ContractWrapperRefund(locksource)
	if err != nil {
		w.logger.Error("WrapperNep5WrapperRefund err", err)
		return CchNeoIssueRetBadParams, "", "", err
	}
	return CchNeoIssueRetOK, txid, "", nil
}

//WrapperNep5GetTxInfo get neo txinfo by txhash
func (w *WrapperServer) WrapperNep5GetTxInfo(txhash string) (result int64, action, fromadd, toaddr string, amount int64, err error) {
	if len(txhash) < WrapperTxHashMinLen {
		w.logger.Error("WrapperNep5GetTxInfo :bad txhash")
		return CchNeoIssueRetBadParams, "", "", "", 0, err
	}
	txinfo, err := w.nta.Nep5GetTxInfo(txhash)
	if err != nil {
		w.logger.Error("WrapperNep5GetTxInfo err", err)
		return CchNeoIssueRetBadParams, "", "", "", 0, err
	}
	return CchNeoIssueRetOK, txinfo.Action, txinfo.Fromaddr, txinfo.Toaddr, txinfo.Amount, nil
}

func (w *WrapperServer) WrapperHashHexStrDelPrix(hexstring string) (string, error) {
	if len(hexstring) == WrapperTxHashMinLen {
		return hexstring, nil
	} else if len(hexstring) == WrapperHashHexStringWithPrix {
		if hexstring[0] == '0' && (hexstring[1] == 'x' || hexstring[1] == 'X') {

			retstring := hexstring[2:]
			return retstring, nil
		}
		return "", errors.New("bad hexstring prefix")
	}
	return "", errors.New("bad hexstring len")
}

func (w *WrapperServer) WrapperHashHexStrAddPrix(hexstring string) (string, error) {
	if len(hexstring) == WrapperTxHashMinLen {
		retstring := "0x"
		retstring = retstring + hexstring
		return retstring, nil
	} else if len(hexstring) == WrapperHashHexStringWithPrix {
		if hexstring[0] == '0' && (hexstring[1] == 'x' || hexstring[1] == 'X') {
			return hexstring, nil
		}
		return "", errors.New("bad hexstring prefix")
	}
	return "", errors.New("bad hexstring len")
}
