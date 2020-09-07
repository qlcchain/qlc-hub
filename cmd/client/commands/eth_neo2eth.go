package commands

import (
	"fmt"

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
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := ethTransaction.WrapperLock(rHash, ethWrapperSignerAddress, 1300000000)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("Wrapper Lock: ", tx)

	b, err := ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if !b || err != nil {
		logger.Fatal(err)
	}
	tx2, err := ethTransaction.UserUnlock(rHash, rOrigin, ethUserAddress)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("User Unlock: ", tx2)
}

func eNeo2EthFetch() {

}
