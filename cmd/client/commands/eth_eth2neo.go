package commands

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"log"
)

func eEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo (not yet timeout)",
		Func: func(c *ishell.Context) {
			eEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func eEth2NeoFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neoFetch",
		Help: "eth -> neo (already timeout)",
		Func: func(c *ishell.Context) {
			eEth2NeoFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func eEth2Neo() {
	var amount int64 = 270000000

	log.Println("=====eth2neo: destroy====")
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := ethTransaction.UserLock(rHash, ethUserAddress, ethWrapperOwnerAddress, amount)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("user lock tx: ", tx)

	b, err := ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}
	tx, err = ethTransaction.WrapperUnlock(rHash, rOrigin, ethWrapperOwnerAddress)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("wrapper unlock tx: ", tx)

	b, err = ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	log.Println("successfully")
}

func eEth2NeoFetch() {
	var amount int64 = 1700000000

	log.Println("=====eth2neo: destroy====")
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := ethTransaction.UserLock(rHash, ethUserAddress, ethWrapperOwnerAddress, amount)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("user lock tx: ", tx)

	b, err := ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	waitingForEthBlocksConfirmed(40)

	tx, err = ethTransaction.UserFetch(rHash, ethUserAddress)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("user fetch: ", tx)

	b, err = ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	log.Println("successfully")
}
