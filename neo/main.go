package main

import (
	"context"
	"log"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

var (
	url             = "http://seed3.ngd.network:20332"
	contractAddress = "b85074ec25aa549814eceb2a4e3748f801c71c51"
	contractUint, _ = util.Uint160DecodeStringBE(contractAddress)
	wif             = ""
	userWif         = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	account, _      = wallet.NewAccountFromWIF(userWif)
)

func main() {
	c, err := client.New(context.Background(), url, client.Options{})
	c.SetWIF(userWif)
	if err != nil {
		log.Fatal(err)
	}

	ps := []request.Param{
		request.Param{
			Type:  request.StringT,
			Value: "userLock",
		},
		request.Param{
			Type:  request.StringT,
			Value: "0xee1445f89e85d41db71debefda3777ce4dbef9bc090fcde91bd906ae403c98c4",
		},
		request.Param{
			Type:  request.StringT,
			Value: "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK",
		},
		request.Param{
			Type:  request.NumberT,
			Value: 10,
		},
		request.Param{
			Type:  request.StringT,
			Value: "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK",
		},
		request.Param{
			Type:  request.NumberT,
			Value: 10,
		},
	}

	scripts, err := request.CreateFunctionInvocationScript(contractUint, ps)
	if err != nil {
		log.Fatal("script ", err)
	}
	tx := transaction.NewInvocationTX(scripts, 1)
	log.Println(tx.Hash())
	err = account.SignTx(tx)
	if err != nil {
		log.Fatal("sign ", err)
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send ", err)
	}
}
