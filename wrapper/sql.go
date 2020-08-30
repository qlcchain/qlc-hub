package wrapper

import (
	"errors"
	_ "fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qlcchain/qlc-hub/log"
	chainctx "github.com/qlcchain/qlc-hub/services/context"
	"go.uber.org/zap"
)

//database table define

//DBNeoMortgageEventTBL  neo nep5 mortgage event
type DBNeoMortgageEventTBL struct {
	ID               int64  `json:"id" pk:"auto" orm:"column(id)"`
	Status           int64  `json:"status" orm:"column(status)"`
	Errno            int64  `json:"error" orm:"column(error)"`
	Amount           int64  `json:"amount" orm:"column(amount)"`
	StartTime        int64  `json:"starttime" orm:"column(starttime)"`
	EndTime          int64  `json:"endtime" orm:"column(endtime)"`
	UserLockNum      int64  `json:"userlocknum" orm:"column(userlocknum)"`
	WrapperLockNum   int64  `json:"wrapperlocknum" orm:"column(wrapperlocknum)"`
	EthLockNum       int64  `json:"ethlocknum" orm:"column(ethlocknum)"`
	NeoLockNum       int64  `json:"neolocknum" orm:"column(neolocknum)"`
	UnlockNum        int64  `json:"unlocknum" orm:"column(unlocknum)"`
	ClaimOutNum      int64  `json:"claimoutnum" orm:"column(claimoutnum)"`
	FetchOutNum      int64  `json:"fetchoutnum" orm:"column(fetchoutnum)"`
	NeoAccount       string `json:"neoaccount" orm:"column(neoaccount);size(128);index"`
	LockHash         string `json:"lockhash" orm:"column(lockhash);size(128);index"`
	HashSource       string `json:"hashsource" orm:"column(hashsource);size(256)"`
	NeoLockTxhash    string `json:"neolocktxhash" orm:"column(neolocktxhash);size(128)"`
	NeoUnlockTxhash  string `json:"neounlocktxhash" orm:"column(neounlocktxhash);size(128)"`
	EthLockTxhash    string `json:"ethlocktxhash" orm:"column(ethlocktxhash);size(128)"`
	EthUnlockTxhash  string `json:"ethunlocktxhash" orm:"column(ethunlocktxhash);size(128)"`
	EthDestoryTxhash string `json:"ethdestorytxhash" orm:"column(ethdestorytxhash);size(128)"`
}

func (m *DBNeoMortgageEventTBL) TableName() string {
	return "neomortgage_event_tbl"
}

//DBEthRedemptionEventTBL  eth erc20 redemption event
type DBEthRedemptionEventTBL struct {
	ID               int64  `json:"id" pk:"auto" orm:"column(id)"`
	Status           int64  `json:"status" orm:"column(status)"`
	Errno            int64  `json:"error" orm:"column(error)"`
	Amount           int64  `json:"amount" orm:"column(amount)"`
	StartTime        int64  `json:"starttime" orm:"column(starttime)"`
	EndTime          int64  `json:"endtime" orm:"column(endtime)"`
	UserLockNum      int64  `json:"userlocknum" orm:"column(userlocknum)"`
	WrapperLockNum   int64  `json:"wrapperlocknum" orm:"column(wrapperlocknum)"`
	EthLockNum       int64  `json:"ethlocknum" orm:"column(ethlocknum)"`
	NeoLockNum       int64  `json:"neolocknum" orm:"column(neolocknum)"`
	UnlockNum        int64  `json:"unlocknum" orm:"column(unlocknum)"`
	ClaimOutNum      int64  `json:"claimoutnum" orm:"column(claimoutnum)"`
	FetchOutNum      int64  `json:"fetchoutnum" orm:"column(fetchoutnum)"`
	EthAccount       string `json:"ethaccount" orm:"column(ethaccount);size(128);index"`
	LockHash         string `json:"lockhash" orm:"column(lockhash);size(128);index"`
	HashSource       string `json:"hashsource" orm:"column(hashsource);size(256)"`
	NeoLockTxhash    string `json:"neolocktxhash" orm:"column(neolocktxhash);size(128)"`
	NeoUnlockTxhash  string `json:"neounlocktxhash" orm:"column(neounlocktxhash);size(128)"`
	EthLockTxhash    string `json:"ethlocktxhash" orm:"column(ethlocktxhash);size(128)"`
	EthUnlockTxhash  string `json:"ethunlocktxhash" orm:"column(ethunlocktxhash);size(128)"`
	EthDestoryTxhash string `json:"ethdestorytxhash" orm:"column(ethdestorytxhash);size(128)"`
}

func (m *DBEthRedemptionEventTBL) TableName() string {
	return "ethredemption_event_tbl"
}

//DBEventStatsChangelogTBL  event status change logs
type DBEventStatsChangelogTBL struct {
	ID        int64  `json:"id" pk:"auto" orm:"column(id)"`
	StatFrom  int64  `json:"stat_from" orm:"column(stat_from)"`
	StatTo    int64  `json:"stat_to" orm:"column(stat_to)"`
	LockHash  string `json:"lockhash" orm:"column(lockhash);size(128);index"`
	Msg       string `json:"msg" orm:"column(msg);size(256)"`
	TimeStamp string `json:"Timestamp" orm:"column(timestamp);type(date);auto_now"`
}

func (m *DBEventStatsChangelogTBL) TableName() string {
	return "event_statchange_logtbl"
}

//DBBlockNumberLogTBL  update block num
type DBBlockNumberLogTBL struct {
	ID         int64  `json:"id" pk:"auto" orm:"column(id)"`
	BlockType  int64  `json:"blocktype" orm:"column(blocktype)"`
	BlockNum   int64  `json:"blocknum" orm:"column(blocknum)"`
	UpdateTime int64  `json:"updatetime" orm:"column(updatetime)"`
	Msg        string `json:"msg" orm:"column(msg)"`
}

func (m *DBBlockNumberLogTBL) TableName() string {
	return "blocknum_logtbl"
}

//DBEthEventLogTBL  eth event log
type DBEthEventLogTBL struct {
	ID         int64  `json:"id" pk:"auto" orm:"column(id)"`
	BlockNum   int64  `json:"blocknum" orm:"column(blocknum)"`
	TimeStamp  int64  `json:"timestamp" orm:"column(timestamp)"`
	VFlag      int64  `json:"vflag" orm:"column(vflag)"`
	Amount     int64  `json:"amount" orm:"column(amount)"`
	LockNum    int64  `json:"locknum" orm:"column(locknum)"`
	UnlockNum  int64  `json:"unlocknum" orm:"column(unlocknum)"`
	Account    string `json:"account" orm:"column(account)"`
	Action     string `json:"action" orm:"column(action)"`
	TxHash     string `json:"txhash" orm:"column(txhash);size(128);index"`
	Lockhash   string `json:"lockhash" orm:"column(lockhash)"`
	HashSource string `json:"hashsource" orm:"column(hashsource)"`
}

func (m *DBEthEventLogTBL) TableName() string {
	return "ethevent_logtbl"
}

//DBNeoEventLogTBL  neo event log
type DBNeoEventLogTBL struct {
	ID         int64  `json:"id" pk:"auto" orm:"column(id)"`
	BlockNum   int64  `json:"blocknum" orm:"column(blocknum)"`
	TimeStamp  int64  `json:"timestamp" orm:"column(timestamp)"`
	VFlag      int64  `json:"vflag" orm:"column(vflag)"`
	Amount     int64  `json:"amount" orm:"column(amount)"`
	From       string `json:"from" orm:"column(from)"`
	To         string `json:"to" orm:"column(to)"`
	Action     string `json:"action" orm:"column(action)"`
	TxHash     string `json:"txhash" orm:"column(txhash);size(128);index"`
	Lockhash   string `json:"lockhash" orm:"column(lockhash)"`
	HashSource string `json:"hashsource" orm:"column(hashsource)"`
}

func (m *DBNeoEventLogTBL) TableName() string {
	return "neoevent_logtbl"
}

//WrapperSqlconn sql connect
type WrapperSqlconn struct {
	logger *zap.SugaredLogger
	ocon   orm.Ormer
}

//NewWrapperSqlconn wrapper sql connect init
func NewWrapperSqlconn() *WrapperSqlconn {
	wsc := &WrapperSqlconn{
		ocon:   orm.NewOrm(),
		logger: log.NewLogger("wrapper sql connect"),
	}
	return wsc
}

//init sql init
func WrapperSqlInit(cfgFile string) {
	cc := chainctx.NewServiceContext(cfgFile)
	cfg, _ := cc.Config()
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		panic(err)
	}
	maxIdle := 30
	maxConn := 50
	dbuser := cfg.SQLCFG.Uname
	dbpwd := cfg.SQLCFG.Upwd
	dburl := cfg.SQLCFG.Url
	dbname := cfg.SQLCFG.DbName
	connect := dbuser + ":" + dbpwd + "@tcp(" + dburl + ")/" + dbname + "?charset=utf8"
	//fmt.Println(connect)
	err = orm.RegisterDataBase("default", "mysql", connect, maxIdle, maxConn)
	if err != nil {
		log.Root.Debugf("RegisterDataBase connect(%s) failed", connect, err)
		return
	}
	orm.RegisterModel(new(DBNeoMortgageEventTBL), new(DBEthRedemptionEventTBL), new(DBEventStatsChangelogTBL), new(DBBlockNumberLogTBL))
	orm.RunSyncdb("default", false, true)
}

//WsqlEventStatChangeLogInsert event status change logrecord insert
func (w *WrapperSqlconn) WsqlEventStatChangeLogInsert(oldstat, newstat int64, lockhash string, msg string) (id int64, err error) {
	var record DBEventStatsChangelogTBL
	record.StatFrom = oldstat
	record.StatTo = newstat
	record.LockHash = lockhash
	record.Msg = msg
	record.TimeStamp = time.Now().Format(WrapperTimestampFormat)
	return w.ocon.Insert(record)
}

//DbEventRecordInsert new event insert
func (w *WrapperSqlconn) DbEventRecordInsert(event *EventInfo) (id int64, err error) {
	if event.Type == cchEventTypeMortgage {
		var mortgageEvent DBNeoMortgageEventTBL
		mortgageEvent.Status = event.Status
		mortgageEvent.Amount = event.Amount
		mortgageEvent.StartTime = event.StartTime
		mortgageEvent.UserLockNum = event.UserLockNum
		mortgageEvent.WrapperLockNum = event.WrapperLockNum
		mortgageEvent.EthLockNum = event.EthLockNum
		mortgageEvent.NeoLockNum = event.NeoLockNum
		mortgageEvent.UnlockNum = event.UnlockNum
		mortgageEvent.ClaimOutNum = event.ClaimOutNum
		mortgageEvent.FetchOutNum = event.FetchOutNum
		mortgageEvent.NeoAccount = event.UserAccount
		mortgageEvent.LockHash = event.LockHash
		mortgageEvent.NeoLockTxhash = event.NeoLockTxhash
		mortgageEvent.EthLockTxhash = event.EthLockTxhash
		return w.ocon.Insert(&mortgageEvent)
	} else if event.Type == cchEventTypeRedemption {
		var redemptionEvent DBEthRedemptionEventTBL
		redemptionEvent.Status = event.Status
		redemptionEvent.Amount = event.Amount
		redemptionEvent.StartTime = event.StartTime
		redemptionEvent.UserLockNum = event.UserLockNum
		redemptionEvent.WrapperLockNum = event.WrapperLockNum
		redemptionEvent.EthLockNum = event.EthLockNum
		redemptionEvent.NeoLockNum = event.NeoLockNum
		redemptionEvent.UnlockNum = event.UnlockNum
		redemptionEvent.ClaimOutNum = event.ClaimOutNum
		redemptionEvent.FetchOutNum = event.FetchOutNum
		redemptionEvent.EthAccount = event.UserAccount
		redemptionEvent.LockHash = event.LockHash
		redemptionEvent.NeoLockTxhash = event.NeoLockTxhash
		redemptionEvent.EthLockTxhash = event.EthLockTxhash
		return w.ocon.Insert(&redemptionEvent)
	}
	return -1, errors.New("bad event.type")
}

//DbGetNeoMortgageEventByLockhash
func (w *WrapperSqlconn) DbGetNeoMortgageEventByLockhash(lockhash string) (info *DBNeoMortgageEventTBL, err error) {
	newinfo := DBNeoMortgageEventTBL{LockHash: lockhash}
	error := w.ocon.Read(&newinfo, "LockHash")
	return &newinfo, error
}

//DbGetEthRedemptionEventByLockhash
func (w *WrapperSqlconn) DbGetEthRedemptionEventByLockhash(lockhash string) (info *DBEthRedemptionEventTBL, err error) {
	newinfo := DBEthRedemptionEventTBL{LockHash: lockhash}
	error := w.ocon.Read(&newinfo, "LockHash")
	return &newinfo, error
}

//DbGetEventByLockhash
func (w *WrapperSqlconn) DbGetEventByLockhash(etype int64, lockhash string) (target *EventInfo, err error) {
	var event EventInfo
	if etype == cchEventTypeMortgage {
		info, err := w.DbGetNeoMortgageEventByLockhash(lockhash)
		if err != nil {
			//w.logger.Debugf("DbGetEventByLockhash:get dbevent by lockhash failed")
			return nil, errors.New("bad lockhash")
		}
		event.Type = etype
		event.Status = info.Status
		event.Amount = info.Amount
		event.StartTime = info.StartTime
		event.UserLockNum = info.UserLockNum
		event.WrapperLockNum = info.WrapperLockNum
		event.EthLockNum = info.EthLockNum
		event.NeoLockNum = info.NeoLockNum
		event.UnlockNum = info.UnlockNum
		event.ClaimOutNum = info.ClaimOutNum
		event.FetchOutNum = info.FetchOutNum
		event.UserAccount = info.NeoAccount
		event.LockHash = info.LockHash
		event.NeoLockTxhash = info.NeoLockTxhash
		event.EthLockTxhash = info.EthLockTxhash
		return &event, nil
	} else if etype == cchEventTypeRedemption {
		info, err := w.DbGetEthRedemptionEventByLockhash(lockhash)
		if err != nil {
			//w.logger.Debugf("DbGetEventByLockhash:get dbevent by lockhash failed")
			return nil, errors.New("bad lockhash")
		}
		event.Type = etype
		event.Status = info.Status
		event.Amount = info.Amount
		event.StartTime = info.StartTime
		event.UserLockNum = info.UserLockNum
		event.WrapperLockNum = info.WrapperLockNum
		event.EthLockNum = info.EthLockNum
		event.NeoLockNum = info.NeoLockNum
		event.UnlockNum = info.UnlockNum
		event.ClaimOutNum = info.ClaimOutNum
		event.FetchOutNum = info.FetchOutNum
		event.UserAccount = info.EthAccount
		event.LockHash = info.LockHash
		event.NeoLockTxhash = info.NeoLockTxhash
		event.EthLockTxhash = info.EthLockTxhash
		return &event, nil
	}
	return nil, errors.New("bad event_type")
}

//DbNeoMortgageUpdate
func (w *WrapperSqlconn) DbNeoMortgageUpdate(info *DBNeoMortgageEventTBL) (id int64, err error) {
	return w.ocon.Update(info)
}

//DbEthRedemptionUpdate
func (w *WrapperSqlconn) DbEthRedemptionUpdate(info *DBEthRedemptionEventTBL) (id int64, err error) {
	return w.ocon.Update(info)
}

//DbEventUpdate
func (w *WrapperSqlconn) DbEventUpdate(event *EventInfo) (id int64, err error) {
	var modifyflag = false
	if event.Type == cchEventTypeMortgage {
		mortgageinfo, err := w.DbGetNeoMortgageEventByLockhash(event.LockHash)
		if err != nil {
			return w.DbEventRecordInsert(event)
		}
		if mortgageinfo.Status != event.Status {
			mortgageinfo.Status = event.Status
			modifyflag = true
		}
		if mortgageinfo.Amount != event.Amount {
			mortgageinfo.Amount = event.Amount
			modifyflag = true
		}
		if mortgageinfo.WrapperLockNum != event.WrapperLockNum {
			mortgageinfo.WrapperLockNum = event.WrapperLockNum
			modifyflag = true
		}
		if mortgageinfo.EthLockNum != event.EthLockNum {
			mortgageinfo.EthLockNum = event.EthLockNum
			modifyflag = true
		}
		if mortgageinfo.NeoLockNum != event.NeoLockNum {
			mortgageinfo.NeoLockNum = event.NeoLockNum
			modifyflag = true
		}
		if mortgageinfo.UnlockNum != event.UnlockNum {
			mortgageinfo.UnlockNum = event.UnlockNum
			modifyflag = true
		}
		if mortgageinfo.ClaimOutNum != event.ClaimOutNum {
			mortgageinfo.ClaimOutNum = event.ClaimOutNum
			modifyflag = true
		}
		if mortgageinfo.FetchOutNum != event.FetchOutNum {
			mortgageinfo.FetchOutNum = event.FetchOutNum
			modifyflag = true
		}
		if mortgageinfo.NeoAccount != event.UserAccount {
			mortgageinfo.NeoAccount = event.UserAccount
			modifyflag = true
		}
		if mortgageinfo.NeoLockTxhash != event.NeoLockTxhash {
			mortgageinfo.NeoLockTxhash = event.NeoLockTxhash
			modifyflag = true
		}
		if mortgageinfo.EthLockTxhash != event.EthLockTxhash {
			mortgageinfo.EthLockTxhash = event.EthLockTxhash
			modifyflag = true
		}
		if modifyflag == true {
			return w.DbNeoMortgageUpdate(mortgageinfo)
		} else {
			return mortgageinfo.ID, nil
		}
	} else if event.Type == cchEventTypeRedemption {
		redeptioninfo, err := w.DbGetEthRedemptionEventByLockhash(event.LockHash)
		if err != nil {
			return w.DbEventRecordInsert(event)
		}
		if redeptioninfo.Status != event.Status {
			redeptioninfo.Status = event.Status
			modifyflag = true
		}
		if redeptioninfo.Amount != event.Amount {
			redeptioninfo.Amount = event.Amount
			modifyflag = true
		}
		if redeptioninfo.WrapperLockNum != event.WrapperLockNum {
			redeptioninfo.WrapperLockNum = event.WrapperLockNum
			modifyflag = true
		}
		if redeptioninfo.EthLockNum != event.EthLockNum {
			redeptioninfo.EthLockNum = event.EthLockNum
			modifyflag = true
		}
		if redeptioninfo.NeoLockNum != event.NeoLockNum {
			redeptioninfo.NeoLockNum = event.NeoLockNum
			modifyflag = true
		}
		if redeptioninfo.UnlockNum != event.UnlockNum {
			redeptioninfo.UnlockNum = event.UnlockNum
			modifyflag = true
		}
		if redeptioninfo.ClaimOutNum != event.ClaimOutNum {
			redeptioninfo.ClaimOutNum = event.ClaimOutNum
			modifyflag = true
		}
		if redeptioninfo.FetchOutNum != event.FetchOutNum {
			redeptioninfo.FetchOutNum = event.FetchOutNum
			modifyflag = true
		}
		if redeptioninfo.EthAccount != event.UserAccount {
			redeptioninfo.EthAccount = event.UserAccount
			modifyflag = true
		}
		if redeptioninfo.NeoLockTxhash != event.NeoLockTxhash {
			redeptioninfo.NeoLockTxhash = event.NeoLockTxhash
			modifyflag = true
		}
		if redeptioninfo.EthLockTxhash != event.EthLockTxhash {
			redeptioninfo.EthLockTxhash = event.EthLockTxhash
			modifyflag = true
		}
		if modifyflag == true {
			return w.DbEthRedemptionUpdate(redeptioninfo)
		} else {
			return redeptioninfo.ID, nil
		}
	}
	return -1, errors.New("bad event type")
}

//WsqlEventDbStatusAndErrnoUpdate db event status and errno update
func (w *WrapperSqlconn) WsqlEventDbStatusAndErrnoUpdate(etype, status, errno int64, lockhash string) (id int64, err error) {
	//w.logger.Debugf("WsqlEventDbStatusAndErrnoUpdate,etype:%d,hash:%s,status：%d，errno:%d", etype, lockhash, status, errno)
	if etype == cchEventTypeMortgage {
		info, err := w.DbGetNeoMortgageEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("WsqlEventDbStatusAndErrnoUpdate:get dbevent by lockhash failed")
			return -1, errors.New("bad lockhash")
		}
		if info.Status != status || info.Errno != errno {
			info.Status = status
			info.Errno = errno
			return w.DbNeoMortgageUpdate(info)
		}
		return info.ID, nil
	} else if etype == cchEventTypeRedemption {
		info, err := w.DbGetEthRedemptionEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("WsqlEventDbStatusAndErrnoUpdate:get dbevent by lockhash failed")
			return -1, errors.New("bad lockhash")
		}
		if info.Status != status || info.Errno != errno {
			info.Status = status
			info.Errno = errno
			return w.DbEthRedemptionUpdate(info)
		}
		return info.ID, nil
	}
	return -1, errors.New("bad event type")
}

//WsqlEventDbStatusUpdate db event status update
func (w *WrapperSqlconn) WsqlEventDbStatusUpdate(etype, status int64, lockhash string) (id int64, err error) {
	//w.logger.Debugf("WsqlEventDbStatusUpdate,etype:%d,hash:%s,status：%d，errno:%d", etype, lockhash, status, errno)
	if etype == cchEventTypeMortgage {
		info, err := w.DbGetNeoMortgageEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("WsqlEventDbStatusUpdate:get dbevent by lockhash failed")
			return -1, errors.New("bad lockhash")
		}
		if info.Status != status {
			info.Status = status
			return w.DbNeoMortgageUpdate(info)
		}
		return info.ID, nil
	} else if etype == cchEventTypeRedemption {
		info, err := w.DbGetEthRedemptionEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("WsqlEventDbStatusUpdate:get dbevent by lockhash failed")
			return -1, errors.New("bad lockhash")
		}
		if info.Status != status {
			info.Status = status
			return w.DbEthRedemptionUpdate(info)
		}
		return info.ID, nil
	}
	return -1, errors.New("bad event type")
}

//DbRunEventNumGetByUser
func (w *WrapperSqlconn) DbRunEventNumGetByUser(etype int64, uaccount string) (num int, err error) {
	var count int
	var sqlcmd string
	if etype == cchEventTypeMortgage {
		sqlcmd = "SELECT count(*) FROM neomortgage_event_tbl where error !=0 AND neoaccount=" + uaccount + ";"
		w.logger.Debugf("sqlcmd(%s)", sqlcmd)
		err := w.ocon.Raw(sqlcmd).QueryRow(&count)
		if err != nil {
			w.logger.Error("DbRunEventNumGetByUser get sql(%s) err(%s)", sqlcmd, err)
			return -1, err
		}
	} else if etype == cchEventTypeRedemption {
		sqlcmd = "SELECT count(*) FROM ethredemption_event_tbl where error !=0 AND ethaccount=" + uaccount + ";"
		w.logger.Debugf("sqlcmd(%s)", sqlcmd)
		err := w.ocon.Raw(sqlcmd).QueryRow(&count)
		if err != nil {
			w.logger.Error("DbRunEventNumGetByUser get sql(%s) err(%s)", sqlcmd, err)
			return -1, err
		}
	} else {
		return -1, errors.New("bad etype")
	}
	return count, nil
}

//WsqlBlockNumberUpdateLogInsert
func (w *WrapperSqlconn) WsqlBlockNumberUpdateLogInsert(btype, blocknum int64, msg string) (id int64, err error) {
	var record DBBlockNumberLogTBL
	record.BlockType = btype
	record.BlockNum = blocknum
	record.Msg = msg
	record.UpdateTime = time.Now().Unix()
	//w.logger.Debugf("WsqlBlockNumberUpdateLogInsert:type(%d) blocknum(%d) msg(%s) ocon(%v)",btype,blocknum,msg,w.ocon)
	return w.ocon.Insert(&record)
}

//WsqlLastBlockNumGet
func (w *WrapperSqlconn) WsqlLastBlockNumGet(btype int64) (int64, error) {
	var count int64
	sqlcmd := "SELECT blocknum from blocknum_logtbl where blocktype=" + strconv.FormatInt(btype, 10) + " ORDER BY ID desc limit 1;"
	//w.logger.Debugf("sqlcmd(%s)", sqlcmd)
	err := w.ocon.Raw(sqlcmd).QueryRow(&count)
	if err != nil {
		w.logger.Error("DbRunEventNumGetByUser get sql(%s) err(%s)", sqlcmd, err)
		return -1, err
	}
	w.logger.Debugf("WsqlLastBlockNumGet get block(%d) num(%d)", btype, count)
	return count, nil
}
