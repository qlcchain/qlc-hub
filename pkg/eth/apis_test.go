package eth

import (
	"testing"
)

const (
	endPoint   = "https://rinkeby.infura.io/v3/0865b420656e4d70bcbbcc76e265fd57"
	endPointws = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	mnemonic   = `lumber choice thing skull allow favorite light horse gun media treat peasant`
	contract   = "0x16e502c867C2d4CAC0F4B4dBd39AB722F5cEc050"

	wrapperPrikey = "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	userEthPrikey = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
)

func getTransaction(t *testing.T) (*Transaction, func()) {
	ethTransaction, err := NewTransaction([]string{endPointws}, contract)
	if err != nil {
		t.Fatal(err)
	}
	return ethTransaction, func() {
	}
}

//func TestGetBestBlockHeight(t *testing.T) {
//	ethTransaction, fn := getTransaction(t)
//	defer fn()
//
//	r, err := ethTransaction.GetBestBlockHeight()
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Log(r)
//}
