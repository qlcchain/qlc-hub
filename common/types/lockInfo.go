package types

import (
	"encoding/json"

	"github.com/qlcchain/qlc-go-sdk/pkg/types"
)

// parse neo transaction result to LockInfo
type LockInfo struct {
	NeoAddress      string        `json:"neoAddress"`
	MultiSigAddress string        `json:"multiSigAddress"`
	QlcAddress      types.Address `json:"qlcAddress"`
	LockTimestamp   int64         `json:"lockTimestamp"`
	UnLockTimestamp int64         `json:"unLockTimestamp"`
	Amount          types.Balance `json:"amount"`
	State           bool          `json:"state"`
}

func (b LockInfo) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}
