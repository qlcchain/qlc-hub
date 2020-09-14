package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"

	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

func nDelete(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "delete",
		Help: "delete swap info",
		Func: func(c *ishell.Context) {
			nDeleteSwapInfo()
		},
	}
	parentCmd.AddCmd(c)
}

func nDeleteSwapInfo() {
	log.Println("====neo delete====")
	rHash := nNeo2Eth()

	w, err := neoTrasaction.QuerySwapInfo(rHash)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(hubUtil.ToString(w))

	log.Println("delete swap info...")
	tx, err := neoTrasaction.DeleteSwapInfo(rHash, neoSignerAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tx)

	time.Sleep(30 * time.Second)

	log.Println("get swap...")
	_, err = neoTrasaction.QuerySwapInfo(rHash)
	if err == nil {
		log.Fatal(err)
	}
}
