package commands

import (
	"fmt"
	"log"

	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
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

	waitForLockerState(rHash, types.WithDrawNeoLockedDone)

	// neo - user unlock
	tx, err = neo.UserUnlock(rOrigin, userWif, neoTrasaction)
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
	waitForLockerState(rHash, types.WithDrawEthUnlockDone)
	logger.Info("successfully")
}
