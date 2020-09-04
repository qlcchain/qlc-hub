package main

import (
	"fmt"
	"log"
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
	contractAddress = "e0abb5fde5a0b870c13f3e60258856e38a939187"
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
	//eth2neo()
	//eth2neoRefund()
}

func neo2eth() {
	log.Println("====neo2eth====")
	n, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := n.UserLock(userWif, wrapperAccount.Address, rHash, 230000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock: ", tx)

	b, _, err := n.TxVerifyAndConfirmed(tx, 1)
	if err != nil {
		log.Fatal(b, err)
	}

	tx, err = n.WrapperUnlock(rOrigin, wrapperWif, userEthAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper unlock: ", tx)
}

func neo2ethRefund() {
	log.Println("====neo2ethRefund====")
	n, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := n.UserLock(userWif, wrapperAccount.Address, rHash, 130000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock: ", tx)
	sleepForHashTimer(40, n)
	n.RefundUser(rOrigin, userWif)
}

func eth2neo() {
	log.Println("====eth2neo====")
	n, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := n.WrapperLock(wrapperWif, userEthAddress, rHash, 140000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper lock: ", tx)

	b, _, err := n.TxVerifyAndConfirmed(tx, 3)
	if err != nil {
		log.Fatal(b, err)
	}

	tx, err = n.UserUnlock(rOrigin, userWif)
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

	tx, err := c.WrapperLock(wrapperWif, userEthAddress, rHash, 140000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper lock: ", tx)

	sleepForHashTimer(20, c)
	refundWrapper("3a985606e258becc169b1bfcb87ce443d9e546f22b0d069fe0cc4caf17afde89", c)
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

func refundWrapper(rHash string, c *neo.Transaction) {
	params := []request.Param{
		neo.FunctionName("refundWrapper"),
		neo.ArrayParams([]request.Param{
			neo.ArrayTypeParam(rHash),
			neo.AddressParam(wrapperAccount.Address),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(neo.TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		RHash:    rHash,
		FuncName: "refundWrapper",
	})
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println("refund wrapper hash ==> ", fmt.Sprintf("0x%s", r))
}
