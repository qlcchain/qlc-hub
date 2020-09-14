package commands

import "github.com/abiosoft/ishell"

func hDelete(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "delete",
		Help: "delete swap info",
		Func: func(c *ishell.Context) {
			hDeleteLockerInfo()
		},
	}
	parentCmd.AddCmd(c)
}

func hDeleteLockerInfo() {

}
