package commands

import (
	"log"
	"time"

	"github.com/abiosoft/ishell"
)

func eDelete(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "delete",
		Help: "delete swap info",
		Func: func(c *ishell.Context) {
			eDeleteLockerInfo()
		},
	}
	parentCmd.AddCmd(c)
}

func eDeleteLockerInfo() {
	log.Println("====eth delete====")

	rHash := eNeo2Eth()

	log.Println("delete hash timer...")
	tx, err := ethTransaction.DeleteHashTimer(rHash, ethOwnerAddress)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("tx confirmed...", tx)
	err = ethTransaction.TxVerifyAndConfirmed(tx, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(30 * time.Second)
	log.Println("get hash timer...")
	e, err := ethTransaction.GetHashTimer(rHash)
	if err != nil {
		log.Fatal(e)
	}

	if e.LockedHeight != 0 || e.UnlockedHeight != 0 {
		log.Fatal(e)
	}
}
