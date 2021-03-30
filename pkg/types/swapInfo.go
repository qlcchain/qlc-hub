package types

import (
	"encoding/json"

	"gorm.io/gorm"
)

type SwapInfo struct {
	gorm.Model
	State          SwapState `msg:"state" json:"state"`
	Chain          ChainType `msg:"chain" json:"chain"`
	Amount         int64     `msg:"amount" json:"amount"`
	EthTxHash      string    `msg:"ethTxHash" json:"ethTxHash"`
	NeoTxHash      string    `msg:"neoTxHash" json:"neoTxHash"`
	EthUserAddr    string    `msg:"ethUserAddr" json:"ethUserAddr"`
	NeoUserAddr    string    `msg:"neoUserAddr" json:"neoUserAddr"`
	StartTime      int64     `msg:"startTime" json:"startTime"`
	LastModifyTime int64     `msg:"lastModifyTime" json:"lastModifyTime"`
	//Interruption   bool      `msg:"interruption" json:"interruption"`
}

func (s *SwapInfo) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

//
//func (b *SwapInfo) Serialize() ([]byte, error) {
//	return b.MarshalMsg(nil)
//}
//
//func (b *SwapInfo) Deserialize(text []byte) error {
//	_, err := b.UnmarshalMsg(text)
//	if err != nil {
//		return err
//	}
//	return nil
//}

type SwapPending struct {
	gorm.Model
	Typ            SwapType  `msg:"typ" json:"typ"`
	Chain          ChainType `msg:"chain" json:"chain"`
	EthTxHash      string    `msg:"ethTxHash" json:"ethTxHash"`
	NeoTxHash      string    `msg:"neoTxHash" json:"neoTxHash"`
	LastModifyTime int64     `msg:"lastModifyTime" json:"lastModifyTime"`
}
