package eth

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/qlcchain/qlc-hub/pkg/util"
)

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	t.Skip()
	ethTransaction, fn := getTransaction(t)
	defer fn()

	if err := ethTransaction.TxVerifyAndConfirmed("0xa3d90416aa98920602ddabdf9f5c3d69e13817aa121e6633f540e6475cf7b0b1", 0, 0); err != nil {
		t.Fatal(err)
	}
}

func TestNewTransaction(t *testing.T) {
	ethTransaction, fn := getTransaction(t)
	defer fn()

	txHash := "0xdde9d2bc6d7ec6c2e78432ed881c3b715cc68e98c58fa8a0c4bb17b0220d77dd"
	tx, p, err := ethTransaction.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
	t.Log(util.ToString(tx))
	fmt.Println("gas: ", tx.Gas())
	fmt.Println("gas price: ", tx.GasPrice())
}

func TestTransaction_GetHashTimer(t *testing.T) {
	ethTransaction, fn := getTransaction(t)
	defer fn()

	txHash := "32fce00156b280b1cf4dd7d0e085a7ab30b1adfb062bfd7bd64a38de290a8817"
	r, err := ethTransaction.GetHashTimer(txHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
