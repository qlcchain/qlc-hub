package wrapper

import (
	"errors"
	_ "fmt"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/services/context"
	"go.uber.org/zap"
	"time"
)

var (
	MortgageEvent   map[string]*EventInfo
	RedemptionEvent map[string]*EventInfo
)

type WrapperServer struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
}

var gWrapperConfig WrapperConfig

//test data
const wrapperNeoAccount string = "ARJZeUehdrFD3Koy3iAymfLDWi3HtCVKYV"
const wrapperNeoPrikey string = "L2pX52zFdqiFyw3kK4rsfPjtcPR9HzSN9h1kE65q9q96BtdAWya8"
const wrapperEthAccount string = "4cD7459d7D228708C090D7d5Dc7ceDF58Cd2cD49"
const wrapperEthPrikey string = "L2pX52zFdqiFyw3kK4rsfPjtcPR9HzSN9h1kE65q9q96BtdAWya8"

//主网合约地址： 0d821bd7b6d53f5c2b40e217c6defc8bbe896cf5
//测试网合约地址： b9d7ea3062e6aeeb3e8ad9548220c4ba1361d263
const wrapperNeoContract string = "b9d7ea3062e6aeeb3e8ad9548220c4ba1361d263"
const wrapperEthContract string = "c7af99fe5513eb6710e6d5f44f9989da40f27f26"
const wrapperEventLimit int64 = 8
const wrapperLockNum int64 = 256

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
	UserAccount      string `json:"useraccount"`
	LockHash         string `json:"lockhash"`
	HashSource       string `json:"hashsource"`
	NeoLockTxhash    string `json:"neolocktxhash"`
	NeoUnlockTxhash  string `json:"neounlocktxhash"`
	EthLockTxhash    string `json:"ethlocktxhash"`
	EthUnlockTxhash  string `json:"ethunlocktxhash"`
	EthDestoryTxhash string `json:"ethdestorytxhash"`
}

//NewWrapperServer wrapper server init
func NewWrapperServer(cfgFile string) *WrapperServer {
	cc := context.NewServiceContext(cfgFile)
	cfg, _ := cc.Config()

	o := &WrapperServer{cfg: cfg}
	o.logger = log.NewLogger("wrapper Server new")

	return o
}

//WrapperEventInit wrapper event init
func (w *WrapperServer) WrapperEventInit() {
	MortgageEvent = make(map[string]*EventInfo)
	RedemptionEvent = make(map[string]*EventInfo)

	//wrapperConfig initialized
	gWrapperConfig.EventLimit = wrapperEventLimit
	gWrapperConfig.LockNum = wrapperLockNum
	gWrapperConfig.NeoAccount = wrapperNeoAccount
	gWrapperConfig.NeoPrikey = wrapperNeoPrikey
	gWrapperConfig.EthAccount = wrapperEthAccount
	gWrapperConfig.EthPrikey = wrapperEthPrikey
	gWrapperConfig.NeoContract = wrapperNeoContract
	gWrapperConfig.EthContract = wrapperEthContract
}

//WrapperEventInsert insert new event node
func (w *WrapperServer) WrapperEventInsert(amount int64, eventType int64, userLocknum int64, lockHash string, txHash string) (err error) {
	var newEvent EventInfo
	if len(txHash) < 32 {
		return errors.New("Bad txHash")
	}
	if len(lockHash) < 32 {
		return errors.New("Bad lockHash")
	}
	newEvent.Type = eventType
	newEvent.LockHash = lockHash
	newEvent.NeoLockTxhash = txHash
	newEvent.Amount = amount
	newEvent.Status = cchEthRedemptionStatusInit
	if eventType == cchEventTypeMortgage {
		MortgageEvent[txHash] = &newEvent
	} else if eventType == cchEventTypeRedemption {
		RedemptionEvent[txHash] = &newEvent
	} else {
		return errors.New("Bad eventType")
	}
	return nil
}

//WrapperEventGetByTxhash get event  by tx_hash
func (w *WrapperServer) WrapperEventGetByTxhash(eventType int64, txHash string) (event *EventInfo, err error) {
	if len(txHash) < 32 {
		return nil, errors.New("Bad txHash")
	}
	if eventType == cchEventTypeMortgage {
		if v, ok := MortgageEvent[txHash]; ok {
			return v, nil
		}
	} else if eventType == cchEventTypeRedemption {
		if v, ok := RedemptionEvent[txHash]; ok {
			return v, nil
		}
	} else {
		return nil, errors.New("Bad eventType")
	}
	return nil, errors.New("no txHash")
}

//WrapperEventUpdateStatByTxhash update event status by tx_hash
func (w *WrapperServer) WrapperEventUpdateStatByTxhash(eventType int64, txHash string, status int64) (err error) {
	if len(txHash) < 32 {
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

func (w *WrapperServer) WrapperOnline() (neoaccount, neocontract, ethaccount, ethcontract string, activetime int64) {
	return gWrapperConfig.NeoAccount, gWrapperConfig.NeoContract, gWrapperConfig.EthAccount, gWrapperConfig.EthContract, time.Now().Unix()
}

func (w *WrapperServer) WrapperNep5LockNotice(eventType, amount int64, txHash, lockHash string) (result int64) {
	if eventType != cchEventTypeMortgage && eventType != cchEventTypeRedemption {
		return cchLockNoticeRetBadParams
	}
	if amount < 0 {
		return cchLockNoticeRetBadParams
	}
	err := w.WrapperEventInsert(amount, eventType, 0, lockHash, txHash)
	if err != nil {
		return cchLockNoticeRetRepeat
	}
	return cchLockNoticeRetOK
}
