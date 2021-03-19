package types

import (
	"github.com/qlcchain/qlc-go-sdk/pkg/types"
	"gorm.io/gorm"
)

type QGasSwapInfo struct {
	gorm.Model
	SwapType       QGasSwapType  `msg:"swapType" json:"swapType"`
	State          QGasSwapState `msg:"state" json:"state"`
	Amount         int64         `msg:"amount" json:"amount"`
	FromAddress    types.Address `msg:"fromAddress" json:"fromAddress"`
	ToAddress      types.Address `msg:"toAddress" json:"toAddress"`
	SendTxHash     types.Hash    `msg:"sendTxHash" json:"sendTxHash"`
	RewardTxHash   types.Hash    `msg:"rewardTxHash" json:"rewardTxHash"`
	EthTxHash      string        `msg:"ethTxHash" json:"ethTxHash"`
	EthUserAddr    string        `msg:"ethUserAddr" json:"ethUserAddr"`
	StartTime      int64         `msg:"startTime" json:"startTime"`
	LastModifyTime int64         `msg:"lastModifyTime" json:"lastModifyTime"`
	//Interruption   bool      `msg:"interruption" json:"interruption"`
}
