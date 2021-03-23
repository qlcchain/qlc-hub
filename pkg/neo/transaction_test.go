package neo

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"

	u "github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	url             = []string{"http://seed5.ngd.network:20332"}
	contractAddress = "bfcbb52d61bc6d3ef2c8cf43f595f4bf5cac66c5"
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
	t.Skip()
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	rHash := "cd76ae68ed900eda74ec78cbc7fd9bc33a9d200546091ba6052f107066e3e66a"

	r, err := c.QuerySwapData(rHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.ToIndentString(r))
	r2, err := c.QueryLockedInfo(rHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.ToIndentString(r2))
}

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	t.Skip()
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	txHash := "789dd4ba43790baf62182b2c3af21d722414a24bcd48a8a2210d06795a4d1a86"
	r, err := c.WaitTxVerifyAndConfirmed(txHash, 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)

	failedHash := "3d2462e274778615d36b7efe493022ec6fd943ccb904a57ec714019e1872fcab"
	r, err = c.WaitTxVerifyAndConfirmed(failedHash, 1)
	if err == nil {
		t.Fatal(r)
	}
	fmt.Println(r)
}

func TestNewTransaction_PublicKey(t *testing.T) {
	account, err := wallet.NewAccountFromWIF("L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR")
	if err != nil {
		t.Fatal(err)
	}
	//pk := account.PrivateKey().PublicKey().String()
	//fmt.Println(pk)
	pks := account.PrivateKey().PublicKey().Bytes()
	pk := hex.EncodeToString(pks)
	fmt.Println(pk)

	pubk, err := keys.NewPublicKeyFromString(pk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(pubk.Bytes()))
}

func TestTransaction_ValidateAddress(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateAddress("ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK"); err != nil {
		t.Fatal(err)
	}
}
