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

	rHash := "8d04db781793f77ce65df25842da51ca7a77fe1ee923bd696382c66588835eae"

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

	r3, d, err := c.lockerEventFromApplicationLog("4cee074f7e2aee185c68d7e3c42035b86a9f3df103396ee8da45cf469bf8a984")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r3, d)
}

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	txHash := "b079aed393f6578a123f4eba5639c8ea3905927444e033d37d986335118395fe"
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
