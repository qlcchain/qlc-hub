package types

import (
	"gorm.io/gorm"
)

type QGasSwapInfo struct {
	gorm.Model
	SwapType       QGasSwapType  `msg:"swapType" json:"swapType"`
	State          QGasSwapState `msg:"state" json:"state"`
	Amount         int64         `msg:"amount" json:"amount"`
	QlcUserAddr    string        `msg:"qlcUserAddr" json:"qlcUserAddr"`
	OwnerAddress   string        `msg:"ownerAddress" json:"ownerAddress"`
	SendTxHash     string        `msg:"sendTxHash" json:"sendTxHash"`
	RewardTxHash   string        `msg:"rewardTxHash" json:"rewardTxHash"`
	UserTxHash     string        `msg:"userTxHash" json:"userTxHash"`
	BlockStr       string        `msg:"blockStr" json:"blockStr"`
	EthTxHash      string        `msg:"ethTxHash" json:"ethTxHash"`
	EthUserAddr    string        `msg:"ethUserAddr" json:"ethUserAddr"`
	StartTime      int64         `msg:"startTime" json:"startTime"`
	LastModifyTime int64         `msg:"lastModifyTime" json:"lastModifyTime"`
	//Interruption   bool      `msg:"interruption" json:"interruption"`
}
