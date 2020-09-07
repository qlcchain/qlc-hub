package commands

import (
	"log"
	"time"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/qlc-hub/pkg/neo"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

func addNeoCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "neo",
		Help: "neo contract",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	nNeo2EthCmd(ethCmd)
	nNeo2EthFetchCmd(ethCmd)
	nEth2NeoCmd(ethCmd)
	nEth2NeoFetchCmd(ethCmd)
}

func nNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth (not yet timeout)",
		Func: func(c *ishell.Context) {
			nNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func nNeo2EthFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethFetch",
		Help: "neo -> eth (already timeout)",
		Func: func(c *ishell.Context) {
			nNeo2EthFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func nNeo2Eth() {
	amount := 220000000

	log.Println("====neo2eth====")
	n, err := neo.NewTransaction(neoUrl, neoContract, singerClient)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := n.UserLock(neoUserAddr, neoWrapperAssetAddr, rHash, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock tx: ", tx)

	b, _, err := n.TxVerifyAndConfirmed(tx, 1)
	if err != nil {
		log.Fatal(b, err)
	}

	tx, err = n.WrapperUnlock(rOrigin, neoWrapperSignerAddress, userEthAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper unlock tx: ", tx)
}

func nNeo2EthFetch() {
	amount := 230000000

	log.Println("====neo2ethRefund====")
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neoTrasaction.UserLock(neoUserAddr, neoWrapperAssetAddr, rHash, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock tx: ", tx)
	waitingForNeoBlocksConfirmed(40)
	tx, err = neoTrasaction.RefundUser(rOrigin, neoWrapperSignerAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user refund tx: ", tx)
}

func waitingForNeoBlocksConfirmed(n uint32) {
	cHeight, err := neoTrasaction.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight
	for {
		time.Sleep(40 * time.Second)
		nHeight, err := neoTrasaction.Client().GetStateHeight()
		if err != nil {
			log.Println(err)
		} else {
			nh := nHeight.BlockHeight
			if nh-ch > n {
				break
			} else {
				log.Printf("waiting for %d/%d block confirmed ... \n", nh-ch, n)
			}
		}
	}
}
