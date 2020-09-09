package commands

import (
	"fmt"
	"log"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

func hEth2NeoCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neo",
		Help: "eth -> neo (not yet timeout)",
		Func: func(c *ishell.Context) {
			hEth2Neo()
		},
	}
	parentCmd.AddCmd(c)
}

func hEth2NeoFetchCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2neoFetch",
		Help: "eth -> neo (already timeout)",
		Func: func(c *ishell.Context) {
			hEth2NeoFetch()
		},
	}
	parentCmd.AddCmd(c)
}

var withdrawAmount = 110000000

func hEth2Neo() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, " ==> ", rHash)

	//  user lock(eth)
	tx, err := ethTransaction.UserLock(rHash, ethUserAddress, ethWrapperOwnerAddress, int64(withdrawAmount))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eth user lock hash: ", tx)

	if !hubWaitingForLockerState(rHash, types.WithDrawNeoLockedDone) {
		log.Fatal(err)
	}

	// eth - wrapper unlock
	paras := fmt.Sprintf(`{
		"rOrigin": "%s",
		"userNep5Addr": "%s"
	}`, rOrigin, neoUserAddr)
	r, err := post(paras, fmt.Sprintf("%s/withdraw/claim", hubUrl))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("claim hash : ", r.(string))

	if !hubWaitingForLockerState(rHash, types.WithDrawEthUnlockDone) {
		log.Fatal(err)
	}
	log.Println("successfully")
}

func hEth2NeoFetch() {
	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, " ==> ", rHash)

	//  user lock(eth)
	tx, err := ethTransaction.UserLock(rHash, ethUserAddress, ethWrapperOwnerAddress, int64(withdrawAmount))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eth user lock hash: ", tx)

	if !hubWaitingForWithdrawEthTimeout(rHash) {
		log.Fatal("timeout")
	}

	tx, err = ethTransaction.UserFetch(rHash, ethUserAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eth user fetch hash: ", tx)
}
