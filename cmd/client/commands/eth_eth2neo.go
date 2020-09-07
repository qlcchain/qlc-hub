package commands

import (
	"github.com/abiosoft/ishell"
)

func eEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo (not yet timeout)",
		Func: func(c *ishell.Context) {
			eEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func eEth2NeoFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neoFetch",
		Help: "eth -> neo (already timeout)",
		Func: func(c *ishell.Context) {
			eEth2NeoFetch()
		},
	}
	parentCmd.AddCmd(c)
}

func eEth2Neo() {

}

func eEth2NeoFetch() {

}
