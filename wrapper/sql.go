package wrapper

import (
	"errors"
	"fmt"
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
	LockBlockNum     int64  `json:"lockblocknum" orm:"column(lockblocknum)"`
	UnlockBlockNum   int64  `json:"unlockblocknum" orm:"column(unlockblocknum)"`
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
	LockBlockNum     int64  `json:"lockblocknum" orm:"column(lockblocknum)"`
	UnlockBlockNum   int64  `json:"unlockblocknum" orm:"column(unlockblocknum)"`
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
	fmt.Println(connect)
	err = orm.RegisterDataBase("default", "mysql", connect, maxIdle, maxConn)
	if err != nil {
		log.Root.Error("RegisterDataBase failed", err)
		return
	}
	orm.RegisterModel(new(DBNeoMortgageEventTBL), new(DBEthRedemptionEventTBL), new(DBEventStatsChangelogTBL))
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

//WsqlEventRecordInsert new event insert
func (w *WrapperSqlconn) WsqlEventRecordInsert(event *EventInfo) (id int64, err error) {
	if event.Type == cchEventTypeMortgage {
		var mortgageEvent DBNeoMortgageEventTBL
		mortgageEvent.Status = event.Status
		mortgageEvent.Amount = event.Amount
		mortgageEvent.StartTime = event.StartTime
		mortgageEvent.UserLockNum = event.UserLockNum
		mortgageEvent.WrapperLockNum = event.WrapperLockNum
		mortgageEvent.LockBlockNum = event.LockBlockNum
		mortgageEvent.UnlockBlockNum = event.UnlockBlockNum
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
		redemptionEvent.LockBlockNum = event.LockBlockNum
		redemptionEvent.UnlockBlockNum = event.UnlockBlockNum
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
			w.logger.Debugf("DbGetEventByLockhash:get dbevent by lockhash failed")
			return nil, errors.New("bad lockhash")
		}
		event.Status = info.Status
		event.Amount = info.Amount
		event.StartTime = info.StartTime
		event.UserLockNum = info.UserLockNum
		event.WrapperLockNum = info.WrapperLockNum
		event.LockBlockNum = info.LockBlockNum
		event.UnlockBlockNum = info.UnlockBlockNum
		event.UserAccount = info.NeoAccount
		event.LockHash = info.LockHash
		event.NeoLockTxhash = info.NeoLockTxhash
		event.EthLockTxhash = info.EthLockTxhash
		return &event, nil
	} else if etype == cchEventTypeRedemption {
		info, err := w.DbGetEthRedemptionEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("DbGetEventByLockhash:get dbevent by lockhash failed")
			return nil, errors.New("bad lockhash")
		}
		event.Status = info.Status
		event.Amount = info.Amount
		event.StartTime = info.StartTime
		event.UserLockNum = info.UserLockNum
		event.WrapperLockNum = info.WrapperLockNum
		event.LockBlockNum = info.LockBlockNum
		event.UnlockBlockNum = info.UnlockBlockNum
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

//WsqlEventDbStatusUpdate db event status update
func (w *WrapperSqlconn) WsqlEventDbStatusUpdate(etype, status, errno int64, lockhash string) (id int64, err error) {
	w.logger.Debugf("WsqlEventDbStatusUpdate,etype:%d,hash:%,status：%d，errno:%d", etype, lockhash, status, errno)
	if etype == cchEventTypeMortgage {
		info, err := w.DbGetNeoMortgageEventByLockhash(lockhash)
		if err != nil {
			w.logger.Debugf("WsqlEventDbStatusUpdate:get dbevent by lockhash failed")
			return -1, errors.New("bad lockhash")
		}
		if info.Status != status {
			info.Status = status
			info.Errno = errno
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
			info.Errno = errno
			return w.DbEthRedemptionUpdate(info)
		}
		return info.ID, nil
	}
	return -1, errors.New("bad event type")
}
