package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

var withdrawAmount = 110000000

func Withdraw() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	logger.Info("hash: ", rOrigin, "==>", rHash)

	// eth - user lock
	_, address, err := eth.GetAccountByPriKey(ethWrapperPrikey)
	if err != nil {
		logger.Fatal(err)
	}
	tx, err := eth.UserLock(rHash, userEthPrikey, address.String(), ethContract, int64(withdrawAmount), ethClient)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("eth user lock hash: ", tx)

	if !waitForLockerState(rHash, types.WithDrawNeoLockedDone) {
		logger.Fatal(err)
	}

	// neo - user unlock
	tx, err = neoTrasaction.UserUnlock(rOrigin, neoUserWif)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo user unlock hash: ", tx)

	// eth - wrapper unlock
	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"rOrigin": "%s",
		"rHash": "%s"
	}`, tx, rOrigin, rHash)
	r, err := post(paras, fmt.Sprintf("%s/withdraw/unlock", hubUrl))
	if !r || err != nil {
		logger.Fatal(err)
	}
	if !waitForLockerState(rHash, types.WithDrawEthUnlockDone) {
		logger.Fatal(err)
	}
	logger.Info("successfully")
}

func WithdrawFetch() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	logger.Info("hash: ", rOrigin, "==>", rHash)

	// eth - user lock
	_, address, err := eth.GetAccountByPriKey(ethWrapperPrikey)
	if err != nil {
		logger.Fatal(err)
	}
	tx, err := eth.UserLock(rHash, userEthPrikey, address.String(), ethContract, int64(withdrawAmount), ethClient)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("eth user lock hash: ", tx)

	waitForNeoIntervalTimerOut(tx)
}

func waitForNeoIntervalTimerOut(txHash string) {
	log.Printf("waiting for timeout  ... \n")
	cHeight, err := neoTrasaction.Client().GetStateHeight()
	if err != nil {
		log.Fatal(err)
	}
	ch := cHeight.BlockHeight

	for i := 0; i < neoIntervalHeight*12; i++ {
		time.Sleep(10 * time.Second)
		b := neoTrasaction.IsConfirmedOverHeightInterval(ch, int64(ethIntervalHeight))
		if b {
			return
		}
	}
	logger.Fatal("timeout ")
}
