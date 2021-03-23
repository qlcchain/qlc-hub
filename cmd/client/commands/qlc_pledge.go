package commands

import (
	"fmt"
	"log" //pb "github.com/qlcchain/qlc-hub/grpc/proto"

	"github.com/abiosoft/ishell"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types" //"github.com/gogo/protobuf/jsonpb"
	//qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
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

	// get pledge send block
	Paras := fmt.Sprintf(`{
		"pledgeAddress":"%s",
		"erc20ReceiverAddr":"%s",
		"amount": "%d"
	}`, qlcUserAddress, "0x73feaa1eE314F8c655E354234017bE2193C9E24E", 100000)
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getPledgeBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	sendHash := result["hash"].(string)
	fmt.Println("result ", result)
	sign, work := signQLCTx(sendHash, result["root"].(string))

	// process send block
	processParas := fmt.Sprintf(`{
		"hash":"%s",
		"signature":"%s",
		"work": "%s"
	}`, sendHash, sign, work)
	pResult, err := post(processParas, fmt.Sprintf("%s/qgasswap/processBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	fmt.Println("process block: ", pResult)
}

func signQLCTx(hash, root string) (string, string) {
	var w qlctypes.Work
	var rootHash qlctypes.Hash
	rootHash.Of(root)
	worker, err := qlctypes.NewWorker(w, rootHash)
	if err != nil {
		log.Fatal(err)
	}
	work := worker.NewWork()

	var blockHash qlctypes.Hash
	blockHash.Of(hash)
	signature := qlcUserAccount.Sign(blockHash)
	return signature.String(), work.String()
}
