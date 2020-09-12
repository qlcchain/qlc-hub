package neo

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	u "github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	url             = "http://seed3.ngd.network:20332"
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

	rHash := "9091cae2c07b6ed45c257341e098d55b3f2924fb83d485804cc927223f214445"

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
	t.Skip()
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	r3, d, err := c.lockerEventFromApplicationLog("4cee074f7e2aee185c68d7e3c42035b86a9f3df103396ee8da45cf469bf8a984")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r3, d)
}

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	t.Skip()
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	txHash := "789dd4ba43790baf62182b2c3af21d722414a24bcd48a8a2210d06795a4d1a86"
	r, err := c.TxVerifyAndConfirmed(txHash, 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)

	failedHash := "3d2462e274778615d36b7efe493022ec6fd943ccb904a57ec714019e1872fcab"
	r, err = c.TxVerifyAndConfirmed(failedHash, 1)
	if err == nil {
		t.Fatal(r)
	}
	fmt.Println(r)
}

func TestTransaction_QlcBalance(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	addr := "Ac2EMY7wCV9Hn9LR1wMWbjgGCqtVofmd6W"
	r, err := c.Balance(addr, "b9d7ea3062e6aeeb3e8ad9548220c4ba1361d263")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}
