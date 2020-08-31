/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"google.golang.org/grpc"

	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	// neo setting
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "0533290f35572cd06e3667653255ffd6ee6430fb"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	userWif           = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _    = wallet.NewAccountFromWIF(userWif)
	wrapperWif        = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _ = wallet.NewAccountFromWIF(wrapperWif)

	userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"

	// eth setting
	ethEndPointws    = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	ethContract      = "0x9a36F711133188EDb3952b3A6ee29c6a3d2e3836"
	ethWrapperPrikey = "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	ethUserPrikey    = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
	userEthPrikey    = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
)

func main() {
	withdraw()
	//deposit()
	//ping()
}

//
//func depositFetch() {
//
//}

func withdraw() {
	neoTrasaction, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	grpcClient, err := grpc.Dial("127.0.0.1:19746", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Close()
	ethClient, err := ethclient.Dial(ethEndPointws)
	if err != nil {
		log.Fatal(err)
	}
	defer ethClient.Close()

	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	//r, err := pb.NewDebugAPIClient(grpcClient).Ping(context.Background(), &empty.Empty{})
	//if err != nil {
	//	log.Fatal(err)
	//}

	// eth user lock
	fmt.Println("======address", wrapperAccount.Address)
	_, address, err := eth.GetAccountByPriKey(ethWrapperPrikey)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := eth.UserLock(rHash, userEthPrikey, address.String(), ethContract, 120000000, ethClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eth user lock hash: ", tx)

	time.Sleep(10 * time.Second)
	waitForLockerState(rHash, types.WithDrawNeoLockedDone, grpcClient)

	tx, err = neo.UserUnlock(rOrigin, userWif, neoTrasaction)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo user unlock hash: ", tx)

	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"rOrigin": "%s",
		"rHash": "%s"
	}`, tx, rOrigin, rHash)
	re, err := post(paras, "http://127.0.0.1:19745/withdraw/unlock")
	if err != nil || !re {
		log.Fatal(re, err)
	}
}

func deposit() {

	neoTrasaction, err := neo.NewTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	grpcClient, err := grpc.Dial("127.0.0.1:19746", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Close()
	ethClient, err := ethclient.Dial(ethEndPointws)
	if err != nil {
		log.Fatal(err)
	}
	defer ethClient.Close()

	rOrigin, rHash := hubUtil.Sha256Hash()
	log.Println("hash: ", rOrigin, "==>", rHash)

	r, err := pb.NewDebugAPIClient(grpcClient).Ping(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.GetNeoAddress(), r.GetEthAddress())

	// user lock (neo)
	tx, err := neo.UserLock(userWif, wrapperAccount.Address, rHash, 130000000, neoTrasaction)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo UserLock hash: ", tx)

	paras := fmt.Sprintf(`{
		"nep5TxHash": "%s",
		"amount": %d,
		"rHash": "%s",
		"addr": "%s"
	}`, tx, 130000000, rHash, wrapperAccount.Address)
	re, err := post(paras, "http://127.0.0.1:19745/deposit/lock")
	if err != nil || !re {
		log.Fatal(re, err)
	}

	waitForLockerState(rHash, types.DepositEthLockedDone, grpcClient)

	// user unlock (eth)
	etx, err := eth.UserUnlock(rHash, rOrigin, ethUserPrikey, ethContract, ethClient)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UserUnlock eth hash: ", etx)
}

func waitForLockerState(rHash string, state types.LockerState, client *grpc.ClientConn) {
	cTicker := time.NewTicker(6 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		er, err := pb.NewDebugAPIClient(client).LockerState(context.Background(), &pb.String{
			Value: rHash,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("====== ", er.StateStr)
		if er.StateStr == types.LockerStateToString(state) {
			return
		}
	}
	log.Fatal("timeout")
}

func ping() {
	jsonStr := []byte("{ }")
	ioBody := bytes.NewBuffer(jsonStr)
	request, err := http.NewRequest("GET", "http://127.0.0.1:19745/debug/ping", ioBody)
	if err != nil {
		log.Fatal("request ", err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("do ", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 200 {
		log.Fatal("code: ", response.StatusCode)
	}

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("do ", err)
	}

	ret := make(map[string]interface{})
	err = json.Unmarshal(bs, &ret)
	if err != nil {
		log.Fatal("do ", err)
	}
	fmt.Println("response: ", string(bs))
}

func post(paras string, url string) (bool, error) {
	jsonStr := []byte(paras)
	ioBody := bytes.NewBuffer(jsonStr)
	request, err := http.NewRequest("POST", url, ioBody)
	if err != nil {
		log.Fatal("request ", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("do ", err)
	}
	defer response.Body.Close()
	if response.StatusCode > 200 {
		log.Fatalf("%d status code returned ", response.StatusCode)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	ret := make(map[string]interface{})
	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		log.Fatal(err)
	}
	if r, ok := ret["value"]; ok != false {
		return r.(bool), nil
	} else {
		if e, ok := ret["error"]; ok != false {
			return false, fmt.Errorf("%s", e)
		}
		return false, fmt.Errorf("response has no result")
	}
}
