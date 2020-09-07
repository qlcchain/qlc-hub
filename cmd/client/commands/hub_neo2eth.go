package commands

import (
	"fmt"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"log"
	"time"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/qlc-hub/pkg/neo"
)

func addHubCmd(shell *ishell.Shell) {
	ethCmd := &ishell.Cmd{
		Name: "hub",
		Help: "hub swap",
		Func: func(c *ishell.Context) {
			c.Println(c.Cmd.HelpText())
		},
	}
	shell.AddCmd(ethCmd)
	hNeo2EthCmd(ethCmd)
	hNeo2EthFetchCmd(ethCmd)
	hEth2NeoCmd(ethCmd)
	hEth2NeoFetchCmd(ethCmd)
}

func hNeo2EthCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2eth",
		Help: "neo -> eth (not yet timeout)",
		Func: func(c *ishell.Context) {
			hNeo2Eth()
		},
	}
	parentCmd.AddCmd(c)
}

func hNeo2EthFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "neo2ethFetch",
		Help: "neo -> eth (already timeout)",
		Func: func(c *ishell.Context) {
			hNeo2EthFetch()
		},
	}
	parentCmd.AddCmd(c)
}

var depositAmount = 130000000

func hNeo2Eth() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, " ==> ", rHash)

	// user lock (neo)
	tx, err := neoTrasaction.UserLock(neoUserAddr, neoWrapperAssetAddr, rHash, 290000000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user lock tx: ", tx)

	// wrapper lock (eth)
	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, lockAmount, rHash, ethWrapperSignerAddress)
	r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	if err != nil || !r {
		log.Fatal(err, r)
	}

	if !waitForLockerState(rHash, types.DepositEthLockedDone) {
		log.Fatal(err)
	}

	// user unlock (eth) -> event -> wrapper unlock (neo)
	etx, err := ethTransaction.UserUnlock(rHash, rOrigin, ethUserAddress)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("UserUnlock eth hash: ", etx)
	if !waitForLockerState(rHash, types.DepositNeoUnLockedDone) {
		log.Fatal(err)
	}
	log.Println("successfully")
}

func hNeo2EthFetch() {
	//rOrigin, rHash := hubUtil.Sha256Hash()
	//logger.Info("hash: ", rOrigin, "==>", rHash)
	//
	//// user lock (neo)
	//tx, err := neoTrasaction.UserLock(neoUserWif, neoWrapperAccount.Address, rHash, depositAmount)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//logger.Info("neo UserLock hash: ", tx)
	//
	//// wrapper lock (eth)
	//paras := fmt.Sprintf(`{
	//	"nep5TxHash": "%s",
	//	"amount": %d,
	//	"rHash": "%s",
	//	"addr": "%s"
	//}`, tx, lockAmount, rHash, ethWrapperAccount.String())
	//r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	//if err != nil || !r {
	//	logger.Fatal(err, r)
	//}
	//
	//// wait for wrapper state
	//if !waitForDepositNeoTimeout(rHash) {
	//	logger.Fatal("timeout")
	//}
	//
	//tx2, err := neoTrasaction.RefundUser(rOrigin, neoUserWif)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	//logger.Info("refund user tx ", tx2)
	//
	//// wrapper lock (eth)
	//paras2 := fmt.Sprintf(`{
	//	"rHash": "%s",
	//	"nep5TxHash": "%s",
	//}`, rHash, tx2)
	//r2, err := post(paras2, fmt.Sprintf("%s/deposit/fetchNotice", hubUrl))
	//if err != nil || !r2 {
	//	logger.Fatal(err, r2)
	//}
}

func waitForEthIntervalTimerOut(rHash string) {
	//log.Printf("waiting for timeout, %s  ... \n", rHash)
	//r, err := getLockerState(rHash)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	//lockerHeight := r["lockedErc20Height"].(float64)
	//
	//for i := 0; i < ethIntervalHeight*12; i++ {
	//	time.Sleep(10 * time.Second)
	//	b, _ := eth.HasConfirmedBlocksHeight(int64(lockerHeight), int64(ethIntervalHeight), ethTransaction)
	//	if b {
	//		return
	//	}
	//}
	//logger.Fatal("timeout ")
}

func sleepForHashTimer(n uint32, c *neo.Transaction) {
	cHeight, err := c.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight
	for {
		time.Sleep(60 * time.Second)
		nHeight, err := c.Client().GetStateHeight()
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
