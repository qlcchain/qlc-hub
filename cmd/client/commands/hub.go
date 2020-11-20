package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/abiosoft/ishell"
)

func addHubCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "hub",
		Help: "hub",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	hNeo2EthCmd(ethCmd)
	hEth2NeoCmd(ethCmd)
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

func hNeo2Eth() {
	amount := 290000000

	// GetNeoUnsignedData
	unsignedParas := fmt.Sprintf(`{
		"amount": %d,
		"erc20ReceiveAddr": "%s"
	}`, amount, ethUserAddress)
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
		"address":"%s",
		"txHash":"%s"
	}`, hex.EncodeToString(sign), hex.EncodeToString(neoUserAccount.PrivateKey().PublicKey().Bytes()), neoUserAddr, neoTxHash)
	r, err = post(sendParas, fmt.Sprintf("%s/deposit/sendNeoTransaction", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}
	log.Println("send neo tx done: ", r["value"].(string))
	time.Sleep(40 * time.Second)

	// GetEthOwnerSign
	ethParas := fmt.Sprintf(`{
		"amount": %d,
		"receiveAddr": "%s",
		"neoTxHash":"%s"
	}`, amount, ethUserAddress, neoTxHash)
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
	fmt.Println("send eth tx done: ", ethTx)
}

func hEth2Neo() {

}
