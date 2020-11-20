package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

func addNeoCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "neo",
		Help: "neo contract",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	nNeo2EthCmd(ethCmd)
	nEth2NeoCmd(ethCmd)
}

func nNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth",
		Func: func(c *ishell.Context) {
			nNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func nEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo",
		Func: func(c *ishell.Context) {
			nEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func waitingForNeoBlocksConfirmed(n uint32) {
	cHeight, err := neoTrasaction.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight
	for {
		time.Sleep(40 * time.Second)
		nHeight, err := neoTrasaction.Client().GetStateHeight()
		if err != nil {
			log.Println(err)
		} else {
			nh := nHeight.BlockHeight
			if nh-ch > n {
				break
			} else {
				log.Printf("waiting for %d/%d block confirmed ... \n", nh-ch, n)
			}
		}
	}
}

func nNeo2Eth() string {
	log.Println("====neo2eth====")
	privateKey := neoUserAccount.PrivateKey()
	publicKey := hex.EncodeToString(privateKey.PublicKey().Bytes())

	amount := 220000000
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)
	rawTx, data, err := neoTrasaction.UnsignedLockTransaction(neoUserAddr, neoAssetAddr, rHash, amount)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("rawTx, ", rawTx)
	log.Println("unsigned data: ", data)

	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	sign := privateKey.Sign(dataBytes)
	tx, err := neoTrasaction.SendLockTransaction(rawTx, hex.EncodeToString(sign), publicKey, neoUserAddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("send tx: ", tx)
	return rHash
}

func nEth2Neo() string {
	return ""
}
