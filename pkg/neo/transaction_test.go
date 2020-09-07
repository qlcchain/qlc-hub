package neo

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"

	u "github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "e0abb5fde5a0b870c13f3e60258856e38a939187"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _     = wallet.NewAccountFromWIF(userWif)
	userAccountUint, _ = address.StringToUint160(userAccount.Address)

	wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)

	userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"
)

func TestNeoTransaction_QuerySwapInfo(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	rHash := "5601c1949d5e6bbc793e4e3ecf62b2b322b8381f71a41b9e1dac04308633fc0c"
	r, err := c.QuerySwapData(rHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.ToIndentString(r))
	r2, err := c.QuerySwapInfo(rHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.ToIndentString(r2))
}

func TestTransaction_RHashFromApplicationLog(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	r3, d, err := c.LockerEventFromApplicationLog("17a561d21f12ca3ad7b98459ccba801fa8ae192c4acdcc5251937fc7dd665566")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r3, d)
}
