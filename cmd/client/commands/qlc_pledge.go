package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
)

func addQLCCmd(shell *ishell.Shell) {
	qlcCmd := &ishell.Cmd{
		Name: "qlc",
		Help: "qlc",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(qlcCmd)
	qQlc2EthCmd(qlcCmd)
	qEth2QlcCmd(qlcCmd)
}

func qQlc2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "qlc2eth",
		Help: "qlc -> eth",
		Func: func(c *ishell.Context) {
			nQlc2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func nQlc2Eth() {

	// GetEthOwnerSign
	qlcPledgeParas := fmt.Sprintf(`{
		"pledgeAddress":"%s",
		"amount": %d,
		"erc20ReceiverAddr":"%s",
	}`, qlcUserAddress, 100, "0x68247e576d0a6128ff655d0fecdea15800f5fcf2")
	bytes, err := postBytes(qlcPledgeParas, fmt.Sprintf("%s/qgasswap/getPledgeBlock", hubUrl))
	if err != nil {
		log.Fatal(err, bytes)
	}

	sendBlk := new(qlctypes.StateBlock)
	err = json.Unmarshal(bytes, &sendBlk)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sendBlk)
}
