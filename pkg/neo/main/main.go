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

	userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _     = wallet.NewAccountFromWIF(userWif)
	userAccountUint, _ = address.StringToUint160(userAccount.Address)

	wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)
)

func main() {
	fmt.Println("user address    ==> ", userAccount.Address)
	fmt.Println("wrapper address ==> ", wrapperAccount.Address)
	rOrigin, rHash := hashValue()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	userLock(rHash)
	//wrapperUnLock(rOrigin)

	//userLockByPkg(rHash)
}

func wrapperUnLock(rOrigin string) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		log.Fatal(err)
	}

	ps := []request.Param{
		{
			Type:  request.StringT,
			Value: "wrapperUnlock",
		}, {
			Type: request.ArrayT,
			Value: []request.Param{
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.StringType,
						Value: request.Param{
							Type:  request.StringT,
							Value: rOrigin,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.ArrayT,
							Value: "ARNpaFJhp6SHziRomrK4cenWw66C8VVFyv",
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
	tx := transaction.NewInvocationTX(scripts, 0)
	//tx.AddVerificationHash(userAccountUint)

	err = wrapperAccount.SignTx(tx)

	//tx.Scripts = append(tx.Scripts, transaction.Witness{
	//	InvocationScript:   append([]byte{byte(opcode.PUSHBYTES64)}, sign...),
	//	VerificationScript: a.getVerificationScript(),
	//})

	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
}

func userLock(rHash string) {
	c, err := client.New(context.Background(), url, client.Options{})
	//c.SetWIF(userWif)
	if err != nil {
		log.Fatal(err)
	}

	fromAddr := hex.EncodeToString(userAccountUint.BytesBE())
	fmt.Println(userAccount.Address, "==>", fromAddr)

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
							Value: rHash,
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
							Value: 110000000,
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
	tx.AddVerificationHash(userAccountUint)
	bys, _ := json.MarshalIndent(tx, "", "\t")
	fmt.Println(string(bys))
	err = userAccount.SignTx(tx)
	if err != nil {
		log.Fatal("sign error: ", err)
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
}

func userLockByPkg(rHash string) {
	client, err := neo.NewNeoTransaction(url, contractAddress)
	if err != nil {
		log.Fatal(err)
	}

	params := []request.Param{
		neo.FunctionName("userLock"),
		neo.ArrayTypeParams([]request.Param{
			neo.HashParam(rHash),
			neo.AddressParam(userAccount.Address),
			neo.IntegerTypeParam(120000000),
			neo.AddressParam(userAccount.Address),
			neo.IntegerTypeParam(rand.Intn(100)),
		}),
	}
	r, err := client.CreateTransaction(params, userWif, 0, 0)
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println(fmt.Sprintf("0x%s", r))
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

func hashValue() (string, string) {
	rOrigin := String(32)
	h := sha256.Sum256([]byte(rOrigin))
	rHash := hex.EncodeToString(h[:])
	return rOrigin, rHash
}
