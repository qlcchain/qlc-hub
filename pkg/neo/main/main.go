package main

import (
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
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
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
	//neo2ethRefund()
	//eth2neo()
	//eth2neoRefund()
}

func neo2eth() {
	log.Println("====neo2eth====")
	c, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neo.UserLock(userWif, wrapperAccount.Address, rHash, 180000000, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock: ", tx)

	b, _, err := neo.TxVerifyAndConfirmed(tx, 3, c)
	if err != nil {
		log.Fatal(b, err)
	}
	tx, err = neo.WrapperUnlock(rOrigin, wrapperWif, userEthAddress, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper unlock: ", tx)
}

func neo2ethRefund() {
	log.Println("====neo2ethRefund====")
	c, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neo.UserLock(userWif, wrapperAccount.Address, rHash, 130000000, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock: ", tx)
	sleepForHashTimer(40, c)
	refundUser(rOrigin, c)
}

func eth2neo() {
	log.Println("====eth2neo====")
	c, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neo.WrapperLock(wrapperWif, userEthAddress, rHash, 140000000, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper lock: ", tx)

	b, _, err := neo.TxVerifyAndConfirmed(tx, 3, c)
	if err != nil {
		log.Fatal(b, err)
	}

	tx, err = neo.UserUnlock(rOrigin, userWif, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user unlock: ", tx)
}

func eth2neoRefund() {
	log.Println("====eth2neoRefund====")
	c, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	wrapperLock(rHash, c)
	sleepForHashTimer(20, c)
	refundWrapper(rOrigin, c)
}

func sleepForHashTimer(n uint32, c *neo.Transaction) {
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

func wrapperLock(rHash string, c *neo.Transaction) {
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

func userUnlock(rOrigin string, c *neo.Transaction) {
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

func refundWrapper(rOrigin string, c *neo.Transaction) {
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

func wrapperUnlock(rOrigin string, c *neo.Transaction) {
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

func refundUser(rOrigin string, c *neo.Transaction) {
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
