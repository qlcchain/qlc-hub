package commands

import (
	"fmt"
	"log"
	"math/big"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

func hEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo",
		Func: func(c *ishell.Context) {
			hEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func hEth2NeoPendingCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neoPending",
		Help: "eth -> neo",
		Func: func(c *ishell.Context) {
			hEth2NeoPending()
		},
	}
	parentCmd.AddCmd(c)
}

func hEth2Neo() {
	amount := 100000000
	ethTx, err := ethTransaction.Burn(ethUserPrivate, neoUserAddr, big.NewInt(int64(amount)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("withdraw send eth tx done: ", ethTx)

	if !waitForSwapState(ethTx, types.SwapStateToString(types.WithDrawDone)) {
		log.Fatal("fail")
	}

	fmt.Println("successfully")
}

func hEth2NeoPending() {
	hash := "0xd9703f6eb3f9f79b97c48f6c5bfcfa445e5d24242dc3b0b8ee54e6ad994b7f41"
	sendParas := fmt.Sprintf(`{
		"hash": "%s",
	}`, hash)
	r, err := post(sendParas, fmt.Sprintf("%s/withdraw/ethTransactionConfirmed", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	fmt.Println("withdraw send eth tx done: ", r["value"].(string))

	if !waitForSwapState(r["value"].(string), types.SwapStateToString(types.WithDrawDone)) {
		log.Fatal("fail")
	}

	fmt.Println("successfully")
}
