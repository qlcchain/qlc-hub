package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

func addHubCmd(shell *ishell.Shell) {
	hubCmd := &ishell.Cmd{
		Name: "hub",
		Help: "hub",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(hubCmd)
	hNeo2EthCmd(hubCmd)
	hNeo2EthByNeoTxCmd(hubCmd)
	hNeo2EthRefundCmd(hubCmd)
	hEth2NeoCmd(hubCmd)
	hEth2NeoPendingCmd(hubCmd)
}

func hNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth",
		Func: func(c *ishell.Context) {
			hNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2EthByNeoTxCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethByNeoTx",
		Help: "neo -> eth",
		Func: func(c *ishell.Context) {
			hNeo2EthByNeoTx()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2EthRefundCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethRefund",
		Help: "neo -> eth",
		Func: func(c *ishell.Context) {
			hNeo2EthRefund()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2EthByNeoTx() {
	amount := 500000000

	//ethUserAddress = "0x2e1ac6242bb084029a9eb29dfb083757d27fced4"
	fmt.Println(ethUserAddress)

	neoTxHash, err := neoTrasaction.CreateLockTransaction(neoUserAddr, ethUserAddress, neoUserWif, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo locked: ", neoTxHash)

	txHash := fmt.Sprintf(`{
		"hash": "%s"
	}`, neoTxHash)
	r, err := post(txHash, fmt.Sprintf("%s/deposit/neoTransactionConfirmed", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForSwapState(neoTxHash, types.SwapStateToString(types.DepositPending)) {
		log.Fatal("fail")
	}

	// GetEthOwnerSign
	ethParas := fmt.Sprintf(`{
		"hash":"%s"
	}`, neoTxHash)
	r, err = post(ethParas, fmt.Sprintf("%s/deposit/getEthOwnerSign", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	ownerSign := r["value"].(string)
	fmt.Println("hub sign: ", ownerSign)

	ethTx, err := ethTransaction.Mint(ethUserPrivate, big.NewInt(int64(amount)), neoTxHash, ownerSign)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("deposit send eth tx done: ", ethTx)

	sentParas := fmt.Sprintf(`{
		"ethTxHash":"%s",
		"neoTxHash":"%s"
	}`, ethTx, neoTxHash)
	r, err = post(sentParas, fmt.Sprintf("%s/deposit/ethTransactionSent", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForSwapState(neoTxHash, types.SwapStateToString(types.DepositDone)) {
		log.Fatal("fail")
	}
	fmt.Println("successfully")
}

func hNeo2Eth() {
	amount := 100000000

	// PackNeoTransaction
	unsignedParas := fmt.Sprintf(`{
		"amount": %d,
		"nep5SenderAddr": "%s",
		"erc20ReceiverAddr": "%s"
	}`, amount, neoUserAddr, ethUserAddress)
	r, err := post(unsignedParas, fmt.Sprintf("%s/deposit/packNeoTransaction", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	neoTxHash := r["txHash"].(string)
	unsignedData := r["unsignedData"].(string)
	log.Println("neo tx hash: ", neoTxHash)
	log.Println("unsigned: ", unsignedData)

	// SendNeoTransaction
	dataBytes, err := hex.DecodeString(unsignedData)
	if err != nil {
		log.Fatal(err)
	}
	sign := neoUserAccount.PrivateKey().Sign(dataBytes)
	sendParas := fmt.Sprintf(`{
		"signature": "%s",
		"publicKey": "%s",
		"nep5SenderAddr":"%s",
		"txHash":"%s"
	}`, hex.EncodeToString(sign), hex.EncodeToString(neoUserAccount.PrivateKey().PublicKey().Bytes()), neoUserAddr, neoTxHash)
	r, err = post(sendParas, fmt.Sprintf("%s/deposit/sendNeoTransaction", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	log.Println("send neo tx done: ", r["value"].(bool))

	if !waitForSwapState(neoTxHash, types.SwapStateToString(types.DepositPending)) {
		log.Fatal("fail")
	}

	// GetEthOwnerSign
	ethParas := fmt.Sprintf(`{
		"hash":"%s"
	}`, neoTxHash)
	r, err = post(ethParas, fmt.Sprintf("%s/deposit/getEthOwnerSign", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	ownerSign := r["value"].(string)
	fmt.Println("hub sign: ", ownerSign)

	ethTx, err := ethTransaction.Mint(ethUserPrivate, big.NewInt(int64(amount)), neoTxHash, ownerSign)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deposit send eth tx done: ", ethTx)

	sentParas := fmt.Sprintf(`{
		"ethTxHash":"%s",
		"neoTxHash":"%s"
	}`, ethTx, neoTxHash)
	r, err = post(sentParas, fmt.Sprintf("%s/deposit/ethTransactionSent", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForSwapState(neoTxHash, types.SwapStateToString(types.DepositDone)) {
		log.Fatal("fail")
	}
	fmt.Println("successfully")
}

func waitForSwapState(hash string, stateStr string) bool {
	cTicker := time.NewTicker(10 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getSwapState(hash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Printf("hash [%s] state is [%s] \n", hash, state["stateStr"])
		if state["stateStr"].(string) == stateStr {
			return true
		}
	}
	log.Fatal("timeout")
	return false
}

func getSwapState(hash string) (map[string]interface{}, error) {
	ret, err := get(fmt.Sprintf("%s/info/swapInfoByTxHash?hash=%s", hubUrl, hash))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func hNeo2EthRefund() {
	amount := 500000000

	//ethUserAddress = "0x2e1ac6242bb084029a9eb29dfb083757d27fced4"
	fmt.Println(ethUserAddress)

	neoTxHash, err := neoTrasaction.CreateLockTransaction(neoUserAddr, ethUserAddress, neoUserWif, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo locked: ", neoTxHash)

	txHash := fmt.Sprintf(`{
		"hash": "%s"
	}`, neoTxHash)
	r, err := post(txHash, fmt.Sprintf("%s/deposit/neoTransactionConfirmed", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForSwapState(neoTxHash, types.SwapStateToString(types.DepositPending)) {
		log.Fatal("fail")
	}

	log.Println("refund...")
	// Refund
	params := fmt.Sprintf(`{
		"hash": "%s"
	}`, neoTxHash)
	r2, err := post(params, fmt.Sprintf("%s/deposit/refund", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	log.Println("refund done: ", r2["value"].(bool))

}
