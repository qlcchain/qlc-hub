package eth

import (
	"fmt"
	"testing"
)

func TestTransaction_SwapInfo(t *testing.T) {
	t.Skip()
	contract := "0xE2484A4178Ce7FfD5cd000030b2a5de08c0Caf8D"
	urls := []string{"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"}
	transaction, err := NewTransaction(urls, contract)
	if err != nil {
		t.Fatal(err)
	}
	neoHash := "0x300a5a08d55ac129896a680b9ec78aa89d23b8f54bda0550868a5c991f519f2c"
	r, err := transaction.GetLockedAmountByNeoTxHash(neoHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}
