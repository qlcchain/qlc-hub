package neo

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
)

func TestTransaction_CreateLockTransaction(t *testing.T) {
	url = []string{"http://seed5.ngd.network:20332"}
	contractAddress = "bfcbb52d61bc6d3ef2c8cf43f595f4bf5cac66c5"
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	userAddr := "AJ5huRnZJj3DZSxnJuZhAMLW1wfc8oMztj"
	//erc20Addr := "0x2e1ac6242bb084029a9eb29dfb083757d27fced4"
	erc20Addr := "0xf6933949C4096670562a5E3a21B8c29c2aacA505"
	wif := "KyiLMuwnkwjNyuQJMmKvmFENCvC4rXAs9BdRSz9HTDmDFt93LRHt"
	amount := 550000000
	tx, err := c.CreateLockTransaction(userAddr, erc20Addr, wif, amount)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tx)
}

func TestTransaction_QuerySwapInfo(t *testing.T) {
	c, err := NewTransaction(url, contractAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	hash := "469718f7cef3ef053231c195960beced5e69e36bce5f659b6a594db8ebfd26e8"

	u, err := util.Uint256DecodeStringLE(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(u.StringLE())
	fmt.Println(u.StringBE())

	s, err := c.QueryLockedInfo(u.StringBE())
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bs))
}
