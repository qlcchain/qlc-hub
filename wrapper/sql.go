package wrapper

import (
	_ "fmt"
	_ "go/doc"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//database table define

//DBNeoMortgageEventTBL  neo nep5 mortgage event
type DBNeoMortgageEventTBL struct {
	ID             		int    `json:"id" pk:"auto" orm:"column(id)"`
	Status          	int    `json:"status" orm:"column(status)"`
	Errno     			int    `json:"error" orm:"column(error)"`
	Amount        		int    `json:"amount" orm:"column(amount)"`
	StartTime       	int    `json:"starttime" orm:"column(starttime)"`
	EndTime    			int    `json:"endtime" orm:"column(endtime)"`
	UserLockNum     	int    `json:"userlocknum" orm:"column(userlocknum)"`
	WrapperLockNum  	int    `json:"wrapperlocknum" orm:"column(wrapperlocknum)"`
	NeoAccount      	string `json:"neoaccount" orm:"column(neoaccount);size(128);index"`
	LockHash        	string `json:"lockhash" orm:"column(lockhash);size(128);index"`
	HashSource      	string `json:"hashsource" orm:"column(hashsource);size(256)"`
	NeoLockTxhash   	string `json:"neolocktxhash" orm:"column(neolocktxhash);size(128)"`
	NeoUnlockTxhash 	string `json:"neounlocktxhash" orm:"column(neounlocktxhash);size(128)"`
	EthLockTxhash   	string `json:"ethlocktxhash" orm:"column(ethlocktxhash);size(128)"`
	EthUnlockTxhash     string `json:"ethunlocktxhash" orm:"column(ethunlocktxhash);size(128)"`
	EthDestoryTxhash 	string `json:"ethdestorytxhash" orm:"column(ethdestorytxhash);size(128)"`
}

func (m *DBNeoMortgageEventTBL) TableName() string {
	return "neomortgage_event_tbl"
}

//DBEthRedemptionEventTBL  eth erc20 redemption event
type DBEthRedemptionEventTBL struct {
	ID             		int    `json:"id" pk:"auto" orm:"column(id)"`
	Status          	int    `json:"status" orm:"column(status)"`
	Errno     			int    `json:"error" orm:"column(error)"`
	Amount        		int    `json:"amount" orm:"column(amount)"`
	StartTime       	int    `json:"starttime" orm:"column(starttime)"`
	EndTime    			int    `json:"endtime" orm:"column(endtime)"`
	UserLockNum     	int    `json:"userlocknum" orm:"column(userlocknum)"`
	WrapperLockNum  	int    `json:"wrapperlocknum" orm:"column(wrapperlocknum)"`
	EthAccount      	string `json:"ethaccount" orm:"column(ethaccount);size(128);index"`
	LockHash        	string `json:"lockhash" orm:"column(lockhash);size(128);index"`
	HashSource      	string `json:"hashsource" orm:"column(hashsource);size(256)"`
	NeoLockTxhash   	string `json:"neolocktxhash" orm:"column(neolocktxhash);size(128)"`
	NeoUnlockTxhash 	string `json:"neounlocktxhash" orm:"column(neounlocktxhash);size(128)"`
	EthLockTxhash   	string `json:"ethlocktxhash" orm:"column(ethlocktxhash);size(128)"`
	EthUnlockTxhash     string `json:"ethunlocktxhash" orm:"column(ethunlocktxhash);size(128)"`
	EthDestoryTxhash 	string `json:"ethdestorytxhash" orm:"column(ethdestorytxhash);size(128)"`
}
func (m *DBEthRedemptionEventTBL) TableName() string {
	return "ethredemption_event_tbl"
}

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	maxIdle := 30
	maxConn := 30
	orm.RegisterDataBase("default", "mysql", "qlcchain:crosschaindb@tcp(127.0.0.1:13306)/wrapper?charset=utf8", maxIdle, maxConn)
	orm.RegisterModel(new(DBNeoMortgageEventTBL), new(DBEthRedemptionEventTBL))
}