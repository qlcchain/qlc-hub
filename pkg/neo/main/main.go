package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/neo"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "0533290f35572cd06e3667653255ffd6ee6430fb"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _     = wallet.NewAccountFromWIF(userWif)
	userAccountUint, _ = address.StringToUint160(userAccount.Address)

	wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)

	userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"
)

func main() {
	neo2eth()
	neo2ethRefund()
	eth2neo()
	eth2neoRefund()
}

func neo2eth() {
	log.Println("====neo2eth====")
	c, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hashValue()
	log.Println("hash: ", rOrigin, "==>", rHash)

	userLock(rHash, c)
	sleepForHashTimer(3, c)
	wrapperUnlock(rOrigin, c)
}

func neo2ethRefund() {
	log.Println("====neo2ethRefund====")
	c, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hashValue()
	log.Println("hash: ", rOrigin, "==>", rHash)

	userLock(rHash, c)
	sleepForHashTimer(40, c)
	refundUser(rOrigin, c)
}

func eth2neo() {
	log.Println("====eth2neo====")
	c, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hashValue()
	log.Println("hash: ", rOrigin, "==>", rHash)

	wrapperLock(rHash, c)
	sleepForHashTimer(3, c)
	userUnlock(rOrigin, c)
}

func eth2neoRefund() {
	log.Println("====eth2neoRefund====")
	c, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hashValue()
	log.Println("hash: ", rOrigin, "==>", rHash)

	wrapperLock(rHash, c)
	sleepForHashTimer(20, c)
	refundWrapper(rOrigin, c)
}

func sleepForHashTimer(n uint32, c *neo.NeoTransaction) {
	log.Printf("waiting for %d block confirmed ... \n", n)
	cHeight, err := c.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight
	for {
		time.Sleep(10 * time.Second)
		nHeight, err := c.Client().GetStateHeight()
		if err != nil {
			log.Println(err)
		} else {
			nh := nHeight.BlockHeight
			if nh-ch > n {
				break
			}
		}
	}
}

func wrapperLock(rHash string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("wrapperLock"),
		neo.ArrayParams([]request.Param{
			neo.ArrayTypeParam(rHash),
			neo.AddressParam(wrapperAccount.Address),
			neo.IntegerTypeParam(140000000),
			neo.ArrayTypeParam(userEthAddress),
			neo.IntegerTypeParam(10),
		}),
	}
	r, err := c.CreateTransaction(neo.TransactionParam{
		Params: params,
		Wif:    wrapperWif,
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("wrapper lock hash ==> ", fmt.Sprintf("0x%s", r))
}

func userUnlock(rOrigin string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("userUnlock"),
		neo.ArrayParams([]request.Param{
			neo.StringTypeParam(rOrigin),
			neo.AddressParam(userAccount.Address),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(neo.TransactionParam{
		Params:   params,
		Wif:      userWif,
		ROrigin:  rOrigin,
		FuncName: "userUnlock",
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("user unlock hash ==> ", fmt.Sprintf("0x%s", r))
}

func refundWrapper(rOrigin string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("refundWrapper"),
		neo.ArrayParams([]request.Param{
			neo.StringTypeParam(rOrigin),
			neo.AddressParam(wrapperAccount.Address),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(neo.TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		ROrigin:  rOrigin,
		FuncName: "refundWrapper",
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("refund wrapper hash ==> ", fmt.Sprintf("0x%s", r))
}

func userLock(rHash string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("userLock"),
		neo.ArrayParams([]request.Param{
			neo.ArrayTypeParam(rHash),
			neo.AddressParam(userAccount.Address),
			neo.IntegerTypeParam(130000000),
			neo.AddressParam(wrapperAccount.Address),
			neo.IntegerTypeParam(10),
		}),
	}
	r, err := c.CreateTransaction(neo.TransactionParam{
		Params: params,
		Wif:    userWif,
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("user lock hash ==> ", fmt.Sprintf("0x%s", r))
}

func wrapperUnlock(rOrigin string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("wrapperUnlock"),
		neo.ArrayParams([]request.Param{
			neo.StringTypeParam(rOrigin),
			neo.AddressParam(wrapperAccount.Address),
			neo.ArrayTypeParam(userEthAddress),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(neo.TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		ROrigin:  rOrigin,
		FuncName: "wrapperUnlock",
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("wrapper unlock hash ==> ", fmt.Sprintf("0x%s", r))
}

func refundUser(rOrigin string, c *neo.NeoTransaction) {
	params := []request.Param{
		neo.FunctionName("refundUser"),
		neo.ArrayParams([]request.Param{
			neo.StringTypeParam(rOrigin),
			neo.AddressParam(userAccount.Address),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(neo.TransactionParam{
		Params:   params,
		Wif:      userWif,
		ROrigin:  rOrigin,
		FuncName: "refundUser",
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("refund user hash ==> ", fmt.Sprintf("0x%s", r))
}

func toString(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "\t")
	return string(data)
}

func remark() []byte {
	remark := make([]byte, 12)
	rand.Read(remark)
	return remark
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func hashValue() (string, string) {
	rOrigin := String(32)
	h := sha256.Sum256([]byte(rOrigin))
	rHash := hex.EncodeToString(h[:])
	return rOrigin, rHash
}
