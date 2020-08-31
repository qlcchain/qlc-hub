/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"google.golang.org/grpc"
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
	ethContract      = "0x6d37597F0d9e917baeF2727ece52AEeb8B5294c7"
	ethWrapperPrikey = "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	ethUserPrikey    = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
	userEthPrikey    = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
)

func main() {
	withdraw()
}

func depositFetch() {

}

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
	tx, err := eth.UserLock(rHash, userEthPrikey, address.String(), ethContract, 220000000, ethClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eth user lock hash: ", tx)

	time.Sleep(40 * time.Second)
	waitForLockerState(rHash, types.WithDrawNeoLockedDone, grpcClient)

	tx, err = neo.UserUnlock(rOrigin, userWif, neoTrasaction)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("neo user unlock hash: ", tx)
	r, err := pb.NewEthAPIClient(grpcClient).WithdrawUnlock(context.Background(), &pb.WithdrawUnlockRequest{
		Nep5TxHash: tx,
		ROrigin:    rOrigin,
		RHash:      rHash,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !r.GetValue() {
		log.Fatal("fail")
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
	fmt.Println(r.GetNep5Address(), r.GetErc20Address())

	// user lock (neo)
	tx, err := neo.UserLock(userWif, wrapperAccount.Address, rHash, 130000000, neoTrasaction)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("UserLock hash: ", tx)
	// wrapper lock (eth)
	er, err := pb.NewEthAPIClient(grpcClient).DepositLock(context.Background(), &pb.DepositLockRequest{
		Nep5TxHash: tx,
		Amount:     130000000,
		RHash:      rHash,
		Addr:       wrapperAccount.Address,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !er.GetValue() {
		log.Fatal("deposit fail")
	}

	waitForLockerState(rHash, types.DepositEthLockedDone, grpcClient)

	// user unlock (eth)
	etx, err := eth.UserUnlock(rHash, rOrigin, ethUserPrikey, ethContract, ethClient)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UserUnlock hash: ", etx)
}

func waitForLockerState(rHash string, state types.LockerState, client *grpc.ClientConn) {
	cTicker := time.NewTicker(6 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		er, err := pb.NewDebugAPIClient(client).LockerState(context.Background(), &pb.String{
			Value: rHash,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("====== ", er.StateStr)
		if er.StateStr == types.LockerStateToString(state) {
			return
		}
	}
	log.Fatal("timeout")
}
