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
	amount := 110000000
	ethTx, err := ethTransactionNep5.Burn(ethUserPrivate, neoUserAddr, big.NewInt(int64(amount)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("withdraw send eth tx done: ", ethTx)

	sentParas := fmt.Sprintf(`{
		"hash":"%s",
		"chainType":"%s"
	}`, ethTx, "eth")
	r, err := post(sentParas, fmt.Sprintf("%s/withdraw/chainTransactionSent", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForSwapState(ethTx, types.SwapStateToString(types.WithDrawDone)) {
		log.Fatal("fail")
	}

	fmt.Println("successfully")
}

func hEth2NeoPending() {
	hash := "0x037260513ec5d8cca4b619d81852608a9a5dcb0e865cb11efcdf2029919738df"
	sendParas := fmt.Sprintf(`{
		"hash": "%s"
	}`, hash)
	r, err := post(sendParas, fmt.Sprintf("%s/withdraw/ethTransactionConfirmed", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	fmt.Println("withdraw send eth tx done: ", r["value"].(bool))

	if !waitForSwapState(hash, types.SwapStateToString(types.WithDrawDone)) {
		log.Fatal("fail")
	}

	fmt.Println("successfully")
}
