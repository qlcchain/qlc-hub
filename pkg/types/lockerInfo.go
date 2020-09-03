package types

//go:generate msgp
type LockerInfo struct {
	State               LockerState `msg:"state" json:"state"`
	RHash               string      `msg:"rHash" json:"rHash"`
	ROrigin             string      `msg:"rOrigin" json:"rOrigin"`
	Amount              int64       `msg:"amount" json:"amount"`
	Erc20Addr           string      `msg:"erc20Addr" json:"erc20Addr"`
	Nep5Addr            string      `msg:"nep5Addr" json:"nep5Addr"`
	LockedNep5Hash      string      `msg:"lnHash" json:"lockedNep5Hash"`
	LockedNep5Height    uint32      `msg:"lnHeight" json:"lockedNep5Height"`
	LockedErc20Hash     string      `msg:"leHash" json:"lockedErc20Hash"`
	LockedErc20Height   uint32      `msg:"leHeight" json:"lockedErc20Height"`
	UnlockedNep5Hash    string      `msg:"unHash" json:"unlockedNep5Hash"`
	UnlockedNep5Height  uint32      `msg:"unHeight" json:"unlockedNep5Height"`
	UnlockedErc20Hash   string      `msg:"ueHash" json:"unlockedErc20Hash"`
	UnlockedErc20Height uint32      `msg:"ueHeight" json:"unlockedErc20Height"`
	StartTime           int64       `msg:"startTime" json:"startTime"`
	LastModifyTime      int64       `msg:"lastModifyTime" json:"lastModifyTime"`
	NeoTimeout          bool        `msg:"neoTimeout" json:"neoTimeout"`
	EthTimeout          bool        `msg:"ethTimeout" json:"ethTimeout"`
	Fail                bool        `msg:"fail" json:"fail"`
	Remark              string      `msg:"remark" json:"remark"`
}

func (b *LockerInfo) Serialize() ([]byte, error) {
	return b.MarshalMsg(nil)
}

func (b *LockerInfo) Deserialize(text []byte) error {
	_, err := b.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}
