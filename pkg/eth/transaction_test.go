package eth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/qlcchain/qlc-hub/pkg/util"
)

//func TestTransaction_TxVerifyAndConfirmed(t *testing.T) {
//	t.Skip()
//	ethTransaction, fn := getTransaction(t)
//	defer fn()
//
//	if err := ethTransaction.WaitTxVerifyAndConfirmed("0xa3d90416aa98920602ddabdf9f5c3d69e13817aa121e6633f540e6475cf7b0b1", 0, 0); err != nil {
//		t.Fatal(err)
//	}
//}

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

func TestTransaction_Sign(t *testing.T) {
	r := make([]byte, 0)
	a := []byte("abc")
	b := []byte("edg")
	r = append(r, a...)
	r = append(r, b...)
	h := sha256.Sum256(r)
	rHash := hex.EncodeToString(h[:])
	fmt.Println(rHash)
}

func TestTransaction_Sign2(t *testing.T) {
	r := make([]byte, 0)

	amount := big.NewInt(0).Mul(big.NewInt(100), big.NewInt(1))

	address := common.HexToAddress("0x5B38Da6a701c568545dCfcB03FcB875f56beddC4")
	a := amount.Bytes()
	b := address.Bytes()
	r = append(r, a...)
	r = append(r, b...)
	fmt.Println(a)
	fmt.Println(hex.EncodeToString(a))
	fmt.Println(b)
	fmt.Println(hex.EncodeToString(b))
	h := sha256.Sum256(r)
	rHash := hex.EncodeToString(h[:])
	fmt.Println("====hash====")
	fmt.Println(rHash)

	r1 := bytes.Repeat([]byte{0}, 32)
	fmt.Println(amount.String())
	fmt.Println(r1)
	copy(r1[len(r1)-len(a):], a)
	fmt.Println(r1)
	fmt.Println(hex.EncodeToString(r1))

	fmt.Println("====hash2====")

	r2 := make([]byte, 0)
	//a2 := amount.Bytes()
	b2 := address.Bytes()
	r2 = append(r2, r1...)
	r2 = append(r2, b2...)

	h2 := sha256.Sum256(r2)
	rHash2 := hex.EncodeToString(h2[:])
	fmt.Println(rHash2)

	fmt.Println("=====sign===")
	privateKey, _, err := GetAccountByPriKey("67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e")
	if err != nil {
		t.Fatal(err)
	}

	sig, err := crypto.Sign(h2[:], privateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(h2[:]))
	fmt.Println(sig)
	fmt.Println(hex.EncodeToString(sig))
}

//func completion(amount *big.Int) []byte {
//	//r := make([]byte,0)
//	//ar := amount.Bytes()
//	//copy[r[0:1], ar]
//
//}

func TestTransaction_SyncLog(t *testing.T) {
	t.Skip()
	urls := []string{"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"}

	contract := "0x0bA64B339281D4F57DF8B535D61c6ceA71CCc956"
	client, err := NewTransaction(urls, contract)
	if err != nil {
		t.Fatal(err)
	}
	hash := "0x98219592dfacebe8988a14a61a13294b1b01fe27dcf31704e29979ca1ec5739e"
	if _, _, _, err := client.SyncBurnLog(hash); err != nil {
		t.Fatal(err)
	}
}

func TestTransaction_Transaction(t *testing.T) {
	t.Skip()
	urls := []string{"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"}
	contract := "0xE2484A4178Ce7FfD5cd000030b2a5de08c0Caf8D"
	signAccount := "aaa052c4f2eed8b96335af467b2ff80dd3a734c57d5ec4b0a8b19e1242ddc601"
	client, err := NewTransaction(urls, contract)
	if err != nil {
		t.Fatal(err)
	}
	instance, opts, err := client.getTransactor(signAccount)
	if err != nil {
		t.Fatal(err)
	}
	recipient := common.HexToAddress("0x255eEcd17E11C5d2FFD5818da31d04B5c1721D7C")
	amount := big.NewInt(20000000000)
	tx, err := instance.Transfer(opts, recipient, amount)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx.Hash().Hex())
}

func TestTransaction_Block(t *testing.T) {
	t.Skip()
	urls1 := []string{"wss://mainnet.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"}
	urls2 := []string{"wss://eth-ws.qlcchain.online"}

	client1, err := NewTransaction(urls1, contract)
	if err != nil {
		t.Fatal(err)
	}

	client2, err := NewTransaction(urls2, contract)
	if err != nil {
		t.Fatal(err)
	}

	hash := common.HexToHash("0xfdb54fad8376f78c1fae5e3afc589cee7668af20d7747f0832590429afc1f7a9")

	_, p, err := client1.client.TransactionByHash(context.Background(), hash)
	t.Log("client1 transactionbyhash: ", p, err)

	_, p, err = client2.client.TransactionByHash(context.Background(), hash)
	t.Log("client2 transactionbyhash: ", p, err)

	_, err = client1.client.TransactionReceipt(context.Background(), hash)
	t.Log("client1 receipt: ", err)

	_, err = client2.client.TransactionReceipt(context.Background(), hash)
	t.Log("client2 receipt: ", err)
}
