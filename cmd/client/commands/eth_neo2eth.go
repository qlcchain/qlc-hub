package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/qlc-hub/pkg/util"
)

func addEthCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "eth",
		Help: "eth contract",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	eNeo2EthCmd(ethCmd)
	eNeo2EthFetchCmd(ethCmd)
	eEth2NeoCmd(ethCmd)
	eEth2NeoFetchCmd(ethCmd)
}

func eNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth (not yet timeout)",
		Func: func(c *ishell.Context) {
			eNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func eNeo2EthFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethFetch",
		Help: "neo -> eth (already timeout)",
		Func: func(c *ishell.Context) {
			eNeo2EthFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func eNeo2Eth() {
	var amount int64 = 1300000000

	log.Println("=====neo2eth: issue====")
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := ethTransaction.WrapperLock(rHash, ethWrapperOwnerAddress, amount)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("Wrapper Lock: ", tx)

	b, err := ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}
	tx, err = ethTransaction.UserUnlock(rHash, rOrigin, ethUserAddress)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("User Unlock: ", tx)

	b, err = ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	log.Println("successfully")
}

func eNeo2EthFetch() {
	var amount int64 = 1300000000

	log.Println("=====neo2eth: issue====")
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := ethTransaction.WrapperLock(rHash, ethWrapperOwnerAddress, amount)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("wrapper Lock: ", tx)

	b, err := ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	waitingForEthBlocksConfirmed(20)

	tx, err = ethTransaction.WrapperFetch(rHash, ethWrapperOwnerAddress)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("wrapper fetch: ", tx)

	b, err = ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}

	log.Println("successfully")
}

func waitingForEthBlocksConfirmed(n int64) {
	cHeight, err := ethTransaction.GetBestBlockHeight()
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(40 * time.Second)
		ch, err := ethTransaction.GetBestBlockHeight()
		if err != nil {
			log.Fatal(err)
		} else {
			if ch-cHeight > n {
				break
			} else {
				log.Printf("waiting for %d/%d block confirmed ... \n", ch-cHeight, n)
			}
		}
	}
}
