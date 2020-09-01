package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

var depositAmount = 130000000

func Deposit() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	logger.Info("hash: ", rOrigin, "==>", rHash)

	// user lock (neo)
	tx, err := neoTrasaction.UserLock(userWif, wrapperAccount.Address, rHash, depositAmount)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("neo UserLock hash: ", tx)

	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, lockAmount, rHash, wrapperAccount.Address)
	r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	if err != nil || !r {
		logger.Fatal(err, r)
	}

	// wait for wrapper state
	waitForLockerState(rHash, types.DepositEthLockedDone)

	// user unlock (eth)
	etx, err := eth.UserUnlock(rHash, rOrigin, ethUserPrikey, ethContract, ethClient)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("UserUnlock eth hash: ", etx)
	waitForLockerState(rHash, types.DepositNeoUnLockedDone)
	logger.Info("successfully")
}

func DepositFetch() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	logger.Info("hash: ", rOrigin, "==>", rHash)

	// user lock (neo)
	tx, err := neoTrasaction.UserLock(userWif, wrapperAccount.Address, rHash, depositAmount)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("neo UserLock hash: ", tx)

	// wrapper lock (eth)
	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, lockAmount, rHash, wrapperAccount.Address)
	r, err := post(paras, fmt.Sprintf("%s/deposit/lock", hubUrl))
	if err != nil || !r {
		logger.Fatal(err, r)
	}

	// wait for wrapper state
	waitForLockerState(rHash, types.DepositEthLockedDone)
	waitForEthIntervalTimerOut(rHash)
}

func waitForEthIntervalTimerOut(rHash string) {
	log.Printf("waiting for timeout, %s  ... \n", rHash)
	r, err := getLockerState(rHash)
	if err != nil {
		logger.Fatal(err)
	}
	lockerHeight := r["lockedErc20Height"].(float64)

	for i := 0; i < ethIntervalHeight*12; i++ {
		time.Sleep(10 * time.Second)
		b := eth.IsConfirmedOverHeightInterval(int64(lockerHeight), int64(ethIntervalHeight), ethClient)
		if b {
			return
		}
	}
	logger.Fatal("timeout ")
}

func sleepForHashTimer(n uint32, c *neo.Transaction) {
	log.Printf("waiting for %d block confirmed ... \n", n)
	cHeight, err := c.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight
	for {
		time.Sleep(10 * time.Second)
		nHeight, err := c.Client().GetStateHeight()
		if err != nil {
			log.Println(err)
		} else {
			nh := nHeight.BlockHeight
			if nh-ch > n {
				break
			}
		}
	}
}
