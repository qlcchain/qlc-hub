package eth

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"testing"
)

func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
	t.Skip()
	ethTransaction, fn := getTransaction(t)
	defer fn()

	if err := ethTransaction.TxVerifyAndConfirmed("0x6aaa384ae047bd9e3f6c5f8cd16a81f3a6d79bf86b0b25447e612346e42cc61e", 0, 0); err != nil {
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
