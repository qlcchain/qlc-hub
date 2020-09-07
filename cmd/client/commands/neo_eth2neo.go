package commands

import (
	"github.com/abiosoft/ishell"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"log"
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
	amount := 140000000

	log.Println("====eth2neo====")
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neoTrasaction.WrapperLock(neoWrapperAssetAddr, userEthAddress, rHash, amount, int(cfg.NEOCfg.WithdrawInterval))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper lock tx: ", tx)

	_, err = neoTrasaction.TxVerifyAndConfirmed(tx, 1)
	if err != nil {
		log.Fatal(err)
	}

	tx, err = neoTrasaction.UserUnlock(rOrigin, neoUserAddr, neoWrapperSignerAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user unlock tx: ", tx)
}

func nEth2NeoFetch() {
	amount := 160000000

	log.Println("====eth2neoRefund====")
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := neoTrasaction.WrapperLock(neoWrapperAssetAddr, userEthAddress, rHash, amount, int(cfg.NEOCfg.WithdrawInterval))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper lock tx: ", tx)

	waitingForNeoBlocksConfirmed(20)

	tx, err = neoTrasaction.RefundWrapper(rHash, neoWrapperSignerAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrapper refund tx: ", tx)
}
