package eth

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	t.Skip()
	client, err := ethclient.Dial(endPointws)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ethTransaction := NewTransaction(client, nil, contract)

	if err := ethTransaction.TxVerifyAndConfirmed("0x6aaa384ae047bd9e3f6c5f8cd16a81f3a6d79bf86b0b25447e612346e42cc61e", 0, 0); err != nil {
		t.Fatal(err)
	}
}
