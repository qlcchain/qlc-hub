package commands

import (
	"fmt"
	"log"

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
	tx, err := neo.UserLock(userWif, wrapperAccount.Address, rHash, depositAmount, neoTrasaction)
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
