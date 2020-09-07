package neo

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	u "github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "278df62f9ba1312f1e1f4b5d239f07beaa1b5b94"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	//userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	//userAccount, _     = wallet.NewAccountFromWIF(userWif)
	//userAccountUint, _ = address.StringToUint160(userAccount.Address)
	//
	//wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	//wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	//wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)
	//
	//userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"
)

func TestNeoTransaction_QuerySwapInfo(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	rHash := "e15c6d4940bbe530f75ccbc06ada1a3674354bed56fe572f6eaf2dd65cf26958"

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
