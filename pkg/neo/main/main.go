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
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/neo"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "0533290f35572cd06e3667653255ffd6ee6430fb"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _     = wallet.NewAccountFromWIF(userWif)
	userAccountUint, _ = address.StringToUint160(userAccount.Address)

	wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)

	userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"
)

func main() {
	fmt.Println("contract ==> ", address.Uint160ToString(contractLE))
	fmt.Println("user address    ==> ", userAccount.Address)
	fmt.Println("wrapper address ==> ", wrapperAccount.Address)
	rOrigin, rHash := hashValue()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	//  q8HrOhrEwpefU673J0HyI0RRh9praomn ==> affb803ca1439f37fae63a6e0dd8bf9e7191260fd9b4725d00a698de1aff9822
	//userLock(rHash)
	//
	wrapperUnLock("q8HrOhrEwpefU673J0HyI0RRh9praomn")
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
	//tx.Attributes = append(tx.Attributes, transaction.Attribute{
	//	Usage: transaction.Script,
	//	Data:  wrapperAccountUint.BytesBE(),
	//})
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  wrapperAccountUint.BytesBE(),
		})

		//r, _ := hex.DecodeString("000001742eafec125f65610c")
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

	d := tx.GetSignedPart()
	fmt.Println(hex.EncodeToString(d))
	sign := wrapperAccount.PrivateKey().Sign(d)
	pack := make([]byte, len(sign)+1)
	pack[0] = byte(len(sign))
	copy(pack[1:], sign)
	fmt.Println(hex.EncodeToString(pack))
	err = wrapperAccount.SignTx(tx)
	fmt.Println("tx", toString(tx))
	fmt.Println(hex.EncodeToString(tx.Bytes()))

	err = c.SendRawTransaction(tx)
	if err != nil {
		log.Fatal("send error: ", err)
	}
	fmt.Println("tx: ", fmt.Sprintf("0x%s", tx.Hash().StringLE()))
	applicationLog(tx.Hash().StringLE(), c)
}

func userLock(rHash string) {
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
							Value: 10,
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

func userLockByPkg(rHash string) {
	c, err := neo.NewNeoTransaction(url, contractAddress)
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
	r, err := c.CreateTransaction(params, userWif, 0, 0)
	if err != nil {
		log.Fatal("tx error: ", err)
	}
	log.Println(fmt.Sprintf("0x%s", r))
}

func applicationLog(hash string, c *client.Client) {
	time.Sleep(30 * time.Second)
	fmt.Println("----application log-----")
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

func toString(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "\t")
	return string(data)
}

func remark() []byte {
	remark := make([]byte, 12)
	rand.Read(remark)
	return remark
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
