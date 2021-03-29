package commands

import (
	"fmt"
	"log" //pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"math/big"

	"github.com/abiosoft/ishell"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types" //"github.com/gogo/protobuf/jsonpb"
	"github.com/qlcchain/qlc-hub/pkg/types"
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
	qQlc2EthPendingCmd(qlcCmd)
	qEth2QlcCmd(qlcCmd)
	qEth2QlcCmdPending(qlcCmd)
	qQlc2BscCmd(qlcCmd)
	qBsc2QlcCmd(qlcCmd)
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

func qQlc2EthPendingCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "qlc2ethPending",
		Help: "qlc -> eth",
		Func: func(c *ishell.Context) {
			nQlc2EthPending()
		},
	}
	parentCmd.AddCmd(c)
}

func nQlc2Eth() {
	amount := 9000000000000000

	// get pledge send block
	Paras := fmt.Sprintf(`{
		"fromAddress":"%s",
		"tokenMintedToAddress":"%s",
		"amount": "%d",
		"chainType": "%s"
	}`, qlcUserAddress, ethUserAddress, amount, "eth")
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getPledgeSendBlock", hubUrl))
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

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgePending)) {
		log.Fatal("fail")
	}

	// GetEthOwnerSign
	signParas := fmt.Sprintf(`{
		"hash":"%s"
	}`, sendHash)
	r, err := post(signParas, fmt.Sprintf("%s/qgasswap/getOwnerSign", hubUrl))
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

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgeDone)) {
		log.Fatal("fail")
	}
	fmt.Println("successfully ")
}

func nQlc2EthPending() {
	amount := 5000000

	sendHash := "93706f493685a34bbff8845c7da51d9ec201d0119c67b51553b7b58d9c439de1"
	sign := "43a190ce947668b5e4d0f8b28da64dcd841a6f65e6cb23e6cfbdff011642e6cb462fd0a43911d1509113819d8fec5d07e19a34ee35a3cbd920f7d28ac0e4cb00"
	work := "00000000002787fe"

	// process send block
	processParas := fmt.Sprintf(`{
		"hash":"%s",
		"signature":"%s",
		"work": "%s"
	}`, sendHash, sign, work)
	pResult, err := post(processParas, fmt.Sprintf("%s/qgasswap/processBlock", hubUrl))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("reward block: ", pResult)

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgePending)) {
		log.Fatal("fail")
	}

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
		log.Fatal(err)
	}
	fmt.Println("reward block: ", sResult)

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgeDone)) {
		log.Fatal("fail")
	}
	fmt.Println("successfully ")
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

func nQlc2Bsc() {
	amount := 9000000000000000

	// get pledge send block
	Paras := fmt.Sprintf(`{
		"fromAddress":"%s",
		"tokenMintedToAddress":"%s",
		"amount": "%d",
		"chainType":"%s"
	}`, qlcUserAddress, bscUserAddress, amount, "bsc")
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getPledgeSendBlock", hubUrl))
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

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgePending)) {
		log.Fatal("fail")
	}

	// GetEthOwnerSign
	signParas := fmt.Sprintf(`{
		"hash":"%s"
	}`, sendHash)
	r, err := post(signParas, fmt.Sprintf("%s/qgasswap/getOwnerSign", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	ownerSign := r["value"].(string)
	fmt.Println("hub sign: ", ownerSign)

	ethTx, err := bscTransactionQLC.QGasMint(bscUserPrivate, big.NewInt(int64(amount)), sendHash, ownerSign)
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

	if !waitForQGasSwapState(sendHash, types.QGasSwapStateToString(types.QGasPledgeDone)) {
		log.Fatal("fail")
	}
	fmt.Println("successfully ")
}

func qQlc2BscCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "qlc2bsc",
		Help: "qlc -> bsc",
		Func: func(c *ishell.Context) {
			nQlc2Bsc()
		},
	}
	parentCmd.AddCmd(c)
}
