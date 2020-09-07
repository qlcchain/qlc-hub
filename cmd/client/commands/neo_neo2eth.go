package commands

import (
	"log"

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
	log.Println("====neo2eth====")
	n, err := neo.NewTransaction(neoUrl, neoContract, singerClient)
	if err != nil {
		log.Fatal(err)
	}
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := n.UserLock(neoUserAddr, neoWrapperSignerAddress, rHash, 230000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock: ", tx)

	b, _, err := n.TxVerifyAndConfirmed(tx, 1)
	if err != nil {
		log.Fatal(b, err)
	}

	//tx, err = n.WrapperUnlock(rOrigin, wrapperWif, userEthAddress)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("wrapper unlock: ", tx)
}

func nNeo2EthFetch() {
	log.Println("====neo2ethRefund====")
	//n, err := neo.NewTransaction(url, contractAddress)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//rOrigin, rHash := hubUtil.Sha256Hash()
	//log.Println("hash: ", rOrigin, "==>", rHash)
	//
	//tx, err := n.UserLock(userWif, wrapperAccount.Address, rHash, 130000000)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("user lock: ", tx)
	//sleepForHashTimer(40, n)
	//n.RefundUser(rOrigin, userWif)
}
