package commands

import (
	"fmt"
	"log" //pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"math/big"

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
	amount := 5000000

	// get pledge send block
	Paras := fmt.Sprintf(`{
		"pledgeAddress":"%s",
		"erc20ReceiverAddr":"%s",
		"amount": "%d"
	}`, qlcUserAddress, ethUserAddress, amount)
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getPledgeBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	sendHash := result["hash"].(string)
	fmt.Println("send Hash: ", sendHash)
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
	fmt.Println("reward block: ", pResult)

	// GetEthOwnerSign
	signParas := fmt.Sprintf(`{
		"hash":"%s"
	}`, sendHash)
	r, err := post(signParas, fmt.Sprintf("%s/qgasswap/getEthOwnerSign", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	ownerSign := r["value"].(string)
	fmt.Println("hub sign: ", ownerSign)

	ethTx, err := ethTransactionQLC.QGasMint(ethUserPrivate, big.NewInt(int64(amount)), sendHash, ownerSign)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("deposit send eth tx done: ", ethTx)

	// process send block
	sentParas := fmt.Sprintf(`{
		"ethTxHash":"%s",
		"qlcTxHash":"%s"
	}`, ethTx, sendHash)
	sResult, err := post(sentParas, fmt.Sprintf("%s/qgasswap/pledgeEthTxSent", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	fmt.Println("reward block: ", sResult)

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
