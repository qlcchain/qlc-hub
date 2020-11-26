package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"
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
	nEth2NeoCmd(ethCmd)
}

func nNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth",
		Func: func(c *ishell.Context) {
			nNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func nEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo",
		Func: func(c *ishell.Context) {
			nEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
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

func nNeo2Eth() string {
	amount := 22000000
	tx, err := neoTrasaction.CreateLockTransaction(neoUserAddr, ethUserAddress, neoUserWif, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lock tx: ", tx)
	return tx
}

func nNeo2EthByHub() string {
	log.Println("====neo2eth====")
	return ""
}

func nEth2Neo() string {
	amount := 20000000

	neoUserAddr := "AJ5huRnZJj3DZSxnJuZhAMLW1wfc8oMztj"
	ethUserAddress = "0x2e1ac6242bb084029a9eb29dfb083757d27fced4"
	ethTxid := "0x51a9de5d8bc002325c0d616ef172ba5c1786580c7837e47606620a920e2eea06"
	tx, err := neoTrasaction.CreateUnLockTransaction(ethTxid, neoUserAddr, ethUserAddress, amount, neoOwnerAddress)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("unlock tx: ", tx)
	return tx
}
