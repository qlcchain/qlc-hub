package commands

import (
	"github.com/abiosoft/ishell"
)

func nEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo (not yet timeout)",
		Func: func(c *ishell.Context) {
			nEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func nEth2NeoFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neoFetch",
		Help: "eth -> neo (already timeout)",
		Func: func(c *ishell.Context) {
			nEth2NeoFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func nEth2Neo() {

}

func nEth2NeoFetch() {

}
