package neo

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"

	u "github.com/qlcchain/qlc-hub/pkg/util"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "0533290f35572cd06e3667653255ffd6ee6430fb"
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
	c, err := NewTransaction(url, contractAddress)
	if err != nil {
		t.Fatal(err)
	}

	rHash := "6c428bbdb7b7a3c235f16d241916337d457b2e52147cb213853f1316aff2e3d3"
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
	//c, err := NewTransaction(url, contractAddress)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//r3, err := c.RHashFromApplicationLog("2493bd842308c4e0e53521099a3a6afd134f55186efd327586c45f0b04c4a21a")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(r3)
}
