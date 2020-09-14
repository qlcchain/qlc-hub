package types

//go:generate msgp
type LockerInfo struct {
	State             LockerState   `msg:"state" json:"state"`
	RHash             string        `msg:"rHash" json:"rHash"`
	ROrigin           string        `msg:"rOrigin" json:"rOrigin"`
	Amount            int64         `msg:"amount" json:"amount"`
	LockedNeoHash     string        `msg:"lnHash" json:"lockedNeoHash"`
	LockedNeoHeight   uint32        `msg:"lnHeight" json:"lockedNeoHeight"`
	LockedEthHash     string        `msg:"leHash" json:"lockedEthHash"`
	LockedEthHeight   uint32        `msg:"leHeight" json:"lockedEthHeight"`
	UnlockedNeoHash   string        `msg:"unHash" json:"unlockedNeoHash"`
	UnlockedNeoHeight uint32        `msg:"unHeight" json:"unlockedNeoHeight"`
	UnlockedEthHash   string        `msg:"ueHash" json:"unlockedEthHash"`
	UnlockedEthHeight uint32        `msg:"ueHeight" json:"unlockedEthHeight"`
	NeoTimerInterval  int64         `msg:"neoTimerInterval" json:"neoTimerInterval"`
	EthTimerInterval  int64         `msg:"ethTimerInterval" json:"ethTimerInterval"`
	NeoUserAddr       string        `msg:"neoUserAddr" json:"neoUserAddr"`
	EthUserAddr       string        `msg:"ethUserAddr" json:"ethUserAddr"`
	GasPrice          int64         `msg:"gasPrice" json:"gasPrice"`
	StartTime         int64         `msg:"startTime" json:"startTime"`
	LastModifyTime    int64         `msg:"lastModifyTime" json:"lastModifyTime"`
	NeoTimeout        bool          `msg:"neoTimeout" json:"neoTimeout"`
	EthTimeout        bool          `msg:"ethTimeout" json:"ethTimeout"`
	Fail              bool          `msg:"fail" json:"fail"`
	Remark            string        `msg:"remark" json:"remark"`
	Interruption      bool          `msg:"interruption" json:"interruption"`
	Deleted           LockerDeleted `msg:"deleted" json:"deleted"`
	DeletedTime       int64         `msg:"deletedTime" json:"deletedTime"`
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
