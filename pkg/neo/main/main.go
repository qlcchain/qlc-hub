package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/neo"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "b85074ec25aa549814eceb2a4e3748f801c71c51"
	contractUint, _ = util.Uint160DecodeStringLE(contractAddress)
	wif             = ""
	userWif         = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	account, _      = wallet.NewAccountFromWIF(userWif)
	from, _         = address.StringToUint160(account.Address)
)

func main() {
	transaction2()
}

func transaction3() {

}

func transaction2() {
	client, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	s := String(10)
	h := sha256.Sum256([]byte(s))
	hash := hex.EncodeToString(h[:])
	fmt.Println(s, "==>", hash)

	params := []request.Param{
		neo.FunctionName("userLock"),
		neo.ArrayTypeParams([]request.Param{
			neo.HashParam(hash),
			neo.AddressParam(account.Address),
			neo.IntegerTypeParam(210000000),
			neo.AddressParam(account.Address),
			neo.IntegerTypeParam(rand.Intn(100)),
		}),
	}
	r, err := client.CreateTransaction(params, userWif, 0, 0)
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println(fmt.Sprintf("0x%s", r))
}

func transaction1() {
	c, err := client.New(context.Background(), url, client.Options{})
	//c.SetWIF(userWif)
	if err != nil {
		log.Fatal(err)
	}

	fromAddr := hex.EncodeToString(from.BytesBE())
	fmt.Println(account.Address, "==>", fromAddr)

	s := String(10)
	h := sha256.Sum256([]byte(s))
	hash := hex.EncodeToString(h[:])
	fmt.Println(s, "==>", hash)

	//pbs, err := util.Uint256DecodeStringBE("42a854d2f9d7f01d4abf03bba5560dc91f3e88d5c71cab17f37872021b247a2d")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//r, err := c.GetRawTransaction(pbs)
	//
	//fmt.Println(r)
	//fmt.Println(err)
	//
	//return

	ps := []request.Param{
		{
			Type:  request.StringT,
			Value: "userLock",
		}, {
			Type: request.ArrayT,
			Value: []request.Param{
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.StringT,
							Value: hash,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.ArrayT,
							Value: fromAddr,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.IntegerType,
						Value: request.Param{
							Type:  request.NumberT,
							Value: 290000000,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.ArrayT,
							Value: fromAddr,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.IntegerType,
						Value: request.Param{
							Type:  request.NumberT,
							Value: 40,
						},
					},
				},
			},
		},
	}

	scripts, err := request.CreateFunctionInvocationScript(contractUint, ps)
	if err != nil {
		log.Fatal("script error: ", err)
	}

	//re, err := c.SignAndPushInvocationTx(scripts, account, 0, 1)
	//if err != nil {
	//	log.Fatal("tx: ", err)
	//}
	//log.Println(re.String())

	tx := transaction.NewInvocationTX(scripts, 0)
	tx.AddVerificationHash(from)
	bys, _ := json.MarshalIndent(tx, "", "\t")
	fmt.Println(string(bys))
	err = account.SignTx(tx)
	if err != nil {
		log.Fatal("sign error: ", err)
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("==tx ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
}

func invoke(client *client.Client) {
	fmt.Println("==== invoke")
	params := []smartcontract.Parameter{
		{
			Type:  smartcontract.ByteArrayType,
			Value: "ee1445f89e85d41db71debefda3777ce4dbef9bc090fcde91bd906ae403c98c4",
		},
		{
			Type:  smartcontract.PublicKeyType,
			Value: "02bfc19e434bb9dde4be76adca4cb39d50bf9832a1ecd347e8a7f6c2bc01a0996f",
		},
		{
			Type:  smartcontract.PublicKeyType,
			Value: "02bfc19e434bb9dde4be76adca4cb39d50bf9832a1ecd347e8a7f6c2bc01a0996f",
		},
		{
			Type:  smartcontract.IntegerType,
			Value: 12,
		},
		{
			Type:  smartcontract.PublicKeyType,
			Value: "02bfc19e434bb9dde4be76adca4cb39d50bf9832a1ecd347e8a7f6c2bc01a0996f",
		},
		{
			Type:  smartcontract.IntegerType,
			Value: 12,
		},
	}
	r, err := client.InvokeFunction(contractAddress, "userLock", params, nil)
	fmt.Println(r)
	fmt.Println(err)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
