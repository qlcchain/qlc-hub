package commands

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"log"
)

func addHubCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "hub",
		Help: "hub swap",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	hNeo2EthCmd(ethCmd)
	hNeo2EthFetchCmd(ethCmd)
	hEth2NeoCmd(ethCmd)
	hEth2NeoFetchCmd(ethCmd)
}

func hNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth (not yet timeout)",
		Func: func(c *ishell.Context) {
			hNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2EthFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethFetch",
		Help: "neo -> eth (already timeout)",
		Func: func(c *ishell.Context) {
			hNeo2EthFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2Eth() {
	amount := 290000000

	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, " ==> ", rHash)

	// user lock (neo)
	tx, err := neoTrasaction.UserLock(neoUserAddr, neoWrapperAssetAddr, rHash, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock tx(neo): ", tx)

	// wrapper lock (eth)
	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, lockAmount, rHash, ethWrapperOwnerAddress)
	r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	if err != nil || !r {
		log.Fatal(err, r)
	}

	if !hubWaitingForLockerState(rHash, types.DepositEthLockedDone) {
		log.Fatal(err)
	}

	// user unlock (eth) -> event -> wrapper unlock (neo)
	etx, err := ethTransaction.UserUnlock(rHash, rOrigin, ethUserAddress)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("user unlock tx(eth): ", etx)
	if !hubWaitingForLockerState(rHash, types.DepositNeoUnLockedDone) {
		log.Fatal(err)
	}
	log.Println("successfully")
}

func hNeo2EthFetch() {
	amount := 290000000

	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, " ==> ", rHash)

	// user lock (neo)
	tx, err := neoTrasaction.UserLock(neoUserAddr, neoWrapperAssetAddr, rHash, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock tx(neo): ", tx)

	// wrapper lock (eth)
	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, lockAmount, rHash, ethWrapperOwnerAddress)
	r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	if err != nil || !r {
		log.Fatal(err, r)
	}

	if !hubWaitingForDepositNeoTimeout(rHash) {
		log.Fatal("timeout")
	}

	// loop/wrapper fetch(eth) -> user fetch (neo)
	paras2 := fmt.Sprintf(`{
		"rOrigin": "%s",
		"userNep5Addr": "%s"
	}`, rOrigin, neoUserAddr)
	r2, err := post(paras2, fmt.Sprintf("%s/deposit/fetch", hubUrl))
	if err != nil || !r2 {
		log.Fatal(err, r2)
	}
}
