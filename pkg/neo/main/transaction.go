package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
)

func neo2ethByNeoGo() {
	rOrigin, rHash := hashValue()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	userLock2(rHash)
	time.Sleep(30 * time.Second)
	wrapperUnlock2(rOrigin)

	//refundUser2(rOrigin)
}

func userLock2(rHash string) {
	c, err := client.New(context.Background(), url, client.Options{})
	//c.SetWIF(userWif)
	if err != nil {
		log.Fatal(err)
	}

	fromAddr := hex.EncodeToString(userAccountUint.BytesBE())
	toAddr := hex.EncodeToString(wrapperAccountUint.BytesBE())
	fmt.Println(userAccount.Address, "==>", fromAddr)
	fmt.Println(wrapperAccount.Address, "==>", toAddr)

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
							Type:  request.ArrayT,
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
							Value: 310000000,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.ArrayT,
							Value: toAddr,
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.IntegerType,
						Value: request.Param{
							Type:  request.NumberT,
							Value: 1,
						},
					},
				},
			},
		},
	}

	scripts, err := request.CreateFunctionInvocationScript(contractLE, ps)
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
	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  userAccountUint.BytesBE(),
		})
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  remark(),
		})
	}

	fmt.Println(toString(tx))
	err = userAccount.SignTx(tx)

	if err != nil {
		log.Fatal("sign error: ", err)
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
	applicationLog(tx.Hash().StringLE(), c)
}

func wrapperUnlock2(rOrigin string) {
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
							Value: hex.EncodeToString(wrapperAccountUint.BytesBE()),
						},
					},
				},
				{
					Type: request.FuncParamT,
					Value: request.FuncParam{
						Type: smartcontract.ByteArrayType,
						Value: request.Param{
							Type:  request.ArrayT,
							Value: userEthAddress,
						},
					},
				},
			},
		},
	}

	scripts, err := request.CreateFunctionInvocationScript(contractLE, ps)
	if err != nil {
		log.Fatal("script error: ", err)
	}

	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(wrapperAccountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  wrapperAccountUint.BytesBE(),
		})

		r := remark()
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	// add witness
	script := io.NewBufBinWriter()
	emit.String(script.BinWriter, rOrigin)
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, "wrapperUnlock")

	tx.Scripts = append(tx.Scripts, transaction.Witness{
		InvocationScript:   script.Bytes(),
		VerificationScript: []byte{},
	})

	//d := tx.GetSignedPart()
	//fmt.Println(hex.EncodeToString(d))
	//sign := wrapperAccount.PrivateKey().Sign(d)
	//pack := make([]byte, len(sign)+1)
	//pack[0] = byte(len(sign))
	//copy(pack[1:], sign)
	//fmt.Println(hex.EncodeToString(pack))
	err = wrapperAccount.SignTx(tx)

	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
	applicationLog(tx.Hash().StringLE(), c)
}

func refundUser2(rOrigin string) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fromAddr := hex.EncodeToString(userAccountUint.BytesBE())
	ps := []request.Param{
		{
			Type:  request.StringT,
			Value: "refundUser",
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
							Value: fromAddr,
						},
					},
				},
			},
		},
	}

	scripts, err := request.CreateFunctionInvocationScript(contractLE, ps)
	if err != nil {
		log.Fatal("script error: ", err)
	}

	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(userAccountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  userAccountUint.BytesBE(),
		})

		r := remark()
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	// add witness
	script := io.NewBufBinWriter()
	emit.String(script.BinWriter, rOrigin)
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, "refundUser")

	tx.Scripts = append(tx.Scripts, transaction.Witness{
		InvocationScript:   script.Bytes(),
		VerificationScript: []byte{},
	})

	err = userAccount.SignTx(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
	applicationLog(tx.Hash().StringLE(), c)
}

func applicationLog(hash string, c *client.Client) {
	time.Sleep(30 * time.Second)
	if h, err := util.Uint256DecodeStringLE(hash); err == nil {
		if l, err := c.GetApplicationLog(h); err == nil {
			data, _ := json.MarshalIndent(l, "", "\t")
			fmt.Println(string(data))
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
}

func querySwapInfo() {
	hash := "d2577e88c208a58fdfd0e1bdb29c1ffee284f9b8f0ba66e5926f8fb34d75fffd"
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	//rsrs, _ := hex.DecodeString(hash)
	params := []smartcontract.Parameter{
		{
			Type:  smartcontract.Hash256Type,
			Value: hash,
		},
	}
	fmt.Println("---", contractLE.StringBE())

	r, err := c.InvokeFunction(contractLE.StringBE(), "querySwapInfo", params, nil)
	fmt.Println(err)
	byset, _ := json.Marshal(r)
	fmt.Println(string(byset))

}
