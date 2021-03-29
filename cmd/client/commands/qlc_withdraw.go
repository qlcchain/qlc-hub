package commands

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

func qEth2QlcCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2qlc",
		Help: "eth -> qlc",
		Func: func(c *ishell.Context) {
			nEth2Qlc()
		},
	}
	parentCmd.AddCmd(c)
}

func qEth2QlcCmdPending(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "eth2qlcPending",
		Help: "eth -> qlc",
		Func: func(c *ishell.Context) {
			nEth2QlcPending()
		},
	}
	parentCmd.AddCmd(c)
}

func nEth2Qlc() {
	amount := 9000000000000000
	ethTx, err := ethTransactionQLC.Burn(ethUserPrivate, qlcUserAddress, big.NewInt(int64(amount)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("withdraw send eth tx done: ", ethTx)

	sentParas := fmt.Sprintf(`{
		"hash":"%s",
		"chainType": "%s"
	}`, ethTx, "eth")
	r, err := post(sentParas, fmt.Sprintf("%s/qgasswap/withdrawEthTxSent", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawPending)) {
		log.Fatal("fail")
	}

	// get withdraw reward block
	Paras := fmt.Sprintf(`{
		"hash":"%s"
	}`, ethTx)
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getWithdrawRewardBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	rewardHash := result["hash"].(string)
	fmt.Println("reward Hash: ", rewardHash)
	sign, work := signQLCTx(rewardHash, result["root"].(string))

	// process send block
	processParas := fmt.Sprintf(`{
		"hash":"%s",
		"signature":"%s",
		"work": "%s"
	}`, rewardHash, sign, work)
	pResult, err := post(processParas, fmt.Sprintf("%s/qgasswap/processBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	fmt.Println("reward block: ", pResult)

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawDone)) {
		log.Fatal("fail")
	}

}

func nEth2QlcPending() {
	ethTx := "0x3e615d7cd90e414b17dbcea238a21ec5c069989084f57be867f94a661ec0bca6"

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawPending)) {
		log.Fatal("fail")
	}

	// get withdraw reward block
	Paras := fmt.Sprintf(`{
		"hash":"%s"
	}`, ethTx)
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getWithdrawRewardBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	rewardHash := result["hash"].(string)
	fmt.Println("reward Hash: ", rewardHash)
	sign, work := signQLCTx(rewardHash, result["root"].(string))

	// process send block
	processParas := fmt.Sprintf(`{
		"hash":"%s",
		"signature":"%s",
		"work": "%s"
	}`, rewardHash, sign, work)
	pResult, err := post(processParas, fmt.Sprintf("%s/qgasswap/processBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	fmt.Println("reward block: ", pResult)

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawDone)) {
		log.Fatal("fail")
	}

}

func waitForQGasSwapState(hash string, stateStr string) bool {
	cTicker := time.NewTicker(10 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getQGasSwapState(hash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Printf("hash [%s] state is [%s] \n", hash, state["stateStr"])
		if state["stateStr"].(string) == stateStr {
			return true
		}
	}
	log.Fatal("timeout")
	return false
}

func getQGasSwapState(hash string) (map[string]interface{}, error) {
	ret, err := get(fmt.Sprintf("%s/qgasswap/swapInfoByTxHash?hash=%s", hubUrl, hash))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func qBsc2QlcCmd(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "bsc2qlc",
		Help: "bsc -> qlc",
		Func: func(c *ishell.Context) {
			nBsc2Qlc()
		},
	}
	parentCmd.AddCmd(c)
}

func nBsc2Qlc() {
	amount := 110000
	ethTx, err := bscTransactionQLC.Burn(bscUserPrivate, qlcUserAddress, big.NewInt(int64(amount)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("withdraw send eth tx done: ", ethTx)

	sentParas := fmt.Sprintf(`{
		"hash":"%s",
		"chainType": "%s"
	}`, ethTx, "bsc")
	r, err := post(sentParas, fmt.Sprintf("%s/qgasswap/withdrawEthTxSent", hubUrl))
	if err != nil {
		log.Fatal(err, r)
	}

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawPending)) {
		log.Fatal("fail")
	}

	// get withdraw reward block
	Paras := fmt.Sprintf(`{
		"hash":"%s"
	}`, ethTx)
	result, err := post(Paras, fmt.Sprintf("%s/qgasswap/getWithdrawBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	rewardHash := result["hash"].(string)
	fmt.Println("reward Hash: ", rewardHash)
	sign, work := signQLCTx(rewardHash, result["root"].(string))

	// process send block
	processParas := fmt.Sprintf(`{
		"hash":"%s",
		"signature":"%s",
		"work": "%s"
	}`, rewardHash, sign, work)
	pResult, err := post(processParas, fmt.Sprintf("%s/qgasswap/processBlock", hubUrl))
	if err != nil {
		log.Fatal(err, result)
	}
	fmt.Println("reward block: ", pResult)

	if !waitForQGasSwapState(ethTx, types.QGasSwapStateToString(types.QGasWithDrawDone)) {
		log.Fatal("fail")
	}

}
