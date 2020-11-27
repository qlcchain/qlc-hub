package eth

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func TestTransaction_SwapInfo(t *testing.T) {
	t.Skip()
	eClient, err := ethclient.Dial("wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57")
	if err != nil {
		t.Fatal(err)
	}
	contract := "0x9d3358268b7cf500766218f152986a5f4ff4d9cc"
	transaction := NewTransaction(eClient, contract)
	neoHash := "94ae1b3068375fe9529c9a3dd9bfb750573b8574b39d84b74c69202274b38f79"
	r, err := transaction.GetLockedAmountByNeoTxHash(neoHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}
