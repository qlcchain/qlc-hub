package commands

import (
	"log"
	"time"

	"github.com/abiosoft/ishell"
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
	//rOrigin, rHash := hubUtil.Sha256Hash()
	//logger.Info("hash: ", rOrigin, "==>", rHash)
	//
	//// eth - user lock
	//_, address, err := eth.GetAccountByPriKey(ethWrapperPrikey)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	//tx, err := eth.UserLock(rHash, ethUserPrikey, address.String(), ethContract, int64(withdrawAmount), ethTransaction)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//logger.Info("eth user lock hash: ", tx)
	//
	//if !waitForLockerState(rHash, types.WithDrawNeoLockedDone) {
	//	logger.Fatal(err)
	//}
	//
	//// neo - user unlock
	//tx, err = neoTrasaction.UserUnlock(rOrigin, neoUserWif)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("neo user unlock hash: ", tx)
	//
	//// eth - wrapper unlock
	//paras := fmt.Sprintf(`{
	//	"nep5TxHash": "%s",
	//	"rOrigin": "%s",
	//	"rHash": "%s"
	//}`, tx, rOrigin, rHash)
	//r, err := post(paras, fmt.Sprintf("%s/withdraw/unlock", hubUrl))
	//if !r || err != nil {
	//	logger.Fatal(err)
	//}
	//if !waitForLockerState(rHash, types.WithDrawEthUnlockDone) {
	//	logger.Fatal(err)
	//}
	//logger.Info("successfully")
}

func hEth2NeoFetch() {
	//rOrigin, rHash := hubUtil.Sha256Hash()
	//logger.Info("hash: ", rOrigin, "==>", rHash)
	//
	//// eth - user lock
	//_, address, err := eth.GetAccountByPriKey(ethWrapperPrikey)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	//tx, err := eth.UserLock(rHash, ethUserPrikey, address.String(), ethContract, int64(withdrawAmount), ethTransaction)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//logger.Info("eth user lock hash: ", tx)
	//
	//if !waitForWithdrawEthTimeout(rHash) {
	//	logger.Fatal("timeout")
	//}
	//
	//tx, err = eth.UserFetch(rHash, ethUserPrikey, ethContract, ethTransaction)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	//logger.Info("eth user fetch hash: ", tx)
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
		b, _ := neoTrasaction.HasConfirmedBlocksHeight(ch, int64(ethIntervalHeight))
		if b {
			return
		}
	}
	logger.Fatal("timeout ")
}
