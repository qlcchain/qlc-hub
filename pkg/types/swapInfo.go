package types

import "gorm.io/gorm"

type SwapInfo struct {
	gorm.Model
	State          SwapState `msg:"state" json:"state"`
	Amount         int64     `msg:"amount" json:"amount"`
	EthTxHash      string    `msg:"ethTxHash" json:"ethTxHash"`
	NeoTxHash      string    `msg:"neoTxHash" json:"neoTxHash"`
	EthUserAddr    string    `msg:"ethUserAddr" json:"ethUserAddr"`
	NeoUserAddr    string    `msg:"neoUserAddr" json:"neoUserAddr"`
	StartTime      int64     `msg:"startTime" json:"startTime"`
	LastModifyTime int64     `msg:"lastModifyTime" json:"lastModifyTime"`
	//Interruption   bool      `msg:"interruption" json:"interruption"`
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
