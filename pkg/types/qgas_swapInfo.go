package types

import (
	"gorm.io/gorm"
)

type QGasSwapInfo struct {
	gorm.Model
	SwapType           QGasSwapType  `msg:"swapType" json:"swapType"`
	State              QGasSwapState `msg:"state" json:"state"`
	Chain              ChainType     `msg:"chain" json:"chain"`
	Amount             int64         `msg:"amount" json:"amount"`
	QlcUserAddr        string        `msg:"qlcUserAddr" json:"qlcUserAddr"`
	OwnerAddress       string        `msg:"ownerAddress" json:"ownerAddress"`
	QlcSendTxHash      string        `msg:"qlcSendTxHash" json:"qlcSendTxHash"`
	QlcRewardTxHash    string        `msg:"qlcRewardTxHash" json:"qlcRewardTxHash"`
	UserTxHash         string        `msg:"userTxHash" json:"userTxHash"`
	BlockStr           string        `msg:"blockStr" json:"blockStr"`
	CrossChainTxHash   string        `msg:"crossChainTxHash" json:"crossChainTxHash"`
	CrossChainUserAddr string        `msg:"crossChainUserAddr" json:"crossChainUserAddr"`
	StartTime          int64         `msg:"startTime" json:"startTime"`
	LastModifyTime     int64         `msg:"lastModifyTime" json:"lastModifyTime"`
	//Interruption   bool      `msg:"interruption" json:"interruption"`
}
