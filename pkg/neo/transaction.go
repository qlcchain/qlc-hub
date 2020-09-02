package neo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
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
	"github.com/qlcchain/qlc-hub/pkg/log"
	u "github.com/qlcchain/qlc-hub/pkg/util"
	"go.uber.org/zap"
)

type Transaction struct {
	url          string
	client       *client.Client
	contractLE   util.Uint160
	contractAddr string
	logger       *zap.SugaredLogger
}

func NewTransaction(url, contractAddr string) (*Transaction, error) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		return nil, err
	}
	contract, err := util.Uint160DecodeStringLE(contractAddr)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		url:          url,
		client:       c,
		contractLE:   contract,
		contractAddr: contractAddr,
		logger:       log.NewLogger("neo/transaction"),
	}, nil
}

type TransactionParam struct {
	Params   []request.Param
	Wif      string
	Sysfee   util.Fixed8
	Netfee   util.Fixed8
	ROrigin  string
	RHash    string
	FuncName string
}

func (n *Transaction) CreateTransaction(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("NewAccountFromWIF: %s", err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("CreateFunctionInvocationScript: %s", err)
	}
	re, err := n.client.SignAndPushInvocationTx(scripts, account, param.Sysfee, param.Netfee)
	if err != nil {
		return "", fmt.Errorf("SignAndPushInvocationTx: %s", err)
	}
	//n.logger.Debugf("transaction successfully: %s", re.StringLE())
	return re.StringLE(), nil
}

type witnessWrapper struct {
	transaction.Witness
	ScriptHash *util.Uint160
}

func (w witnessWrapper) GetScriptHash() *util.Uint160 {
	if w.ScriptHash == nil {
		hash := w.Witness.ScriptHash()
		w.ScriptHash = &hash
	}
	return w.ScriptHash
}

func (n *Transaction) CreateTransactionAppendWitness(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("new account: %s", err)
	}
	accountUint, err := address.StringToUint160(account.Address)
	if err != nil {
		return "", fmt.Errorf("address uint160: %s", err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("create script: %s", err)
	}
	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(accountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  n.contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  accountUint.BytesBE(),
		})
		r := remark()
		//r, _ := hex.DecodeString("00000174483c1ff2d76670ab")
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	if err := account.SignTx(tx); err != nil {
		return "", fmt.Errorf("signTx: %s", err)
	}

	var tmp []*witnessWrapper
	// sign
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   tx.Scripts[0].InvocationScript,
			VerificationScript: tx.Scripts[0].VerificationScript,
		},
		ScriptHash: nil,
	})
	// add witness
	script := io.NewBufBinWriter()
	if param.ROrigin != "" && param.RHash == "" {
		emit.String(script.BinWriter, param.ROrigin)
	} else if param.ROrigin == "" && param.RHash != "" {
		rHex, err := hex.DecodeString(param.RHash)
		if err != nil {
			return "", fmt.Errorf("decode error: %s", err)
		}
		emit.Bytes(script.BinWriter, rHex)
	} else {
		return "", errors.New("invalid r text")
	}
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, param.FuncName)
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   script.Bytes(),
			VerificationScript: []byte{},
		},
		ScriptHash: &n.contractLE,
	})

	sort.Slice(tmp, func(i, j int) bool {
		h1 := tmp[i].GetScriptHash()
		h2 := tmp[j].GetScriptHash()

		return big.NewInt(0).SetBytes(h1.BytesLE()).Cmp(big.NewInt(0).SetBytes(h2.BytesLE())) < 0
	})

	size := len(tmp)
	witness := make([]transaction.Witness, size)
	for i := 0; i < size; i++ {
		witness[i] = transaction.Witness{
			InvocationScript:   tmp[i].InvocationScript,
			VerificationScript: tmp[i].VerificationScript,
		}
	}

	tx.Scripts = witness

	n.logger.Debug(hex.EncodeToString(tx.Bytes()))
	n.logger.Debug(u.ToIndentString(tx))

	if err := n.client.SendRawTransaction(tx); err != nil {
		return "", fmt.Errorf("sendRawTransaction: %s", err)
	}
	//n.logger.Debugf("transaction successfully: %s", tx.Hash().StringLE())
	return tx.Hash().StringLE(), nil
}

func remark() []byte {
	remark := make([]byte, 12)
	rand.Read(remark)
	return remark
}

func (n *Transaction) Client() *client.Client {
	return n.client
}

func (n *Transaction) QuerySwapData(rHash string) (map[string]interface{}, error) {
	hash, err := hex.DecodeString(rHash)
	if err != nil {
		return nil, err
	}
	params := []smartcontract.Parameter{
		{
			Type:  smartcontract.ByteArrayType,
			Value: hash,
		},
	}
	r, err := n.client.InvokeFunction(n.contractAddr, "querySwapInfo", params, nil)
	if err != nil {
		return nil, err
	} else if r.State != "HALT" || len(r.Stack) == 0 {
		return nil, errors.New("invalid VM state")
	}

	return StackToSwapInfo(r.Stack)
}

type SwapInfo struct {
	Amount   int64
	UserAddr string
	rHash    string
	rOrigin  string
	OverTime int64
}

func (n *Transaction) QuerySwapInfo(rHash string) (*SwapInfo, error) {
	data, err := n.QuerySwapData(rHash)
	if err != nil {
		return nil, err
	}
	info := new(SwapInfo)
	amount, err := getAmount(data)
	if err != nil {
		return nil, err
	}
	info.Amount = amount
	overTime, err := getIntValue("overtimeBlocks", data)
	if err != nil {
		return nil, err
	}
	info.OverTime = overTime
	//origin, err := getStringValue("origin", data)
	//if err != nil {
	//	return nil, err
	//}
	//info.rOrigin = origin
	return info, nil
}

func getAmount(data map[string]interface{}) (int64, error) {
	amount, err := getValue("amount", data)
	if err != nil {
		return 0, err
	}
	if r, ok := amount.(*big.Int); ok {
		return r.Int64(), nil
	} else {
		return 0, errors.New("invalid amount")
	}
}

func getIntValue(key string, data map[string]interface{}) (int64, error) {
	if v, err := getValue(key, data); err != nil {
		return 0, err
	} else {
		if r, ok := v.(int64); ok {
			return r, nil
		} else {
			return 0, errors.New("invalid string")
		}
	}
}

func getStringValue(key string, data map[string]interface{}) (string, error) {
	if v, err := getValue(key, data); err != nil {
		return "", err
	} else {
		if r, ok := v.(string); ok {
			return r, nil
		} else {
			return "", errors.New("invalid string")
		}
	}
}

func getValue(key string, data map[string]interface{}) (interface{}, error) {
	if r, ok := data[key]; ok {
		return r, nil
	} else {
		return 0, fmt.Errorf("can not get key %s [%s]", key, u.ToIndentString(data))
	}
}

func (n *Transaction) TxVerifyAndConfirmed(txHash string, interval int) (bool, uint32, error) {
	//todo verify tx successfully

	var txHeight uint32
	cTicker := time.NewTicker(6 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			hash, err := util.Uint256DecodeStringLE(txHash)
			if err != nil {
				return false, 0, fmt.Errorf("tx verify decode hash: %s", err)
			}
			txHeight, err = n.client.GetTransactionHeight(hash)
			if err != nil {
				fmt.Println("======= ", txHash, err)
			} else {
				goto HeightConfirmed
			}
		case <-cTimer.C:
			return false, 0, fmt.Errorf("neo tx by hash timeout: %s", txHash)
		}
	}

HeightConfirmed:
	nTicker := time.NewTicker(6 * time.Second)
	nTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-nTicker.C:
			nHeight, err := n.Client().GetStateHeight()
			if err != nil {
				return false, 0, err
			} else {
				nh := nHeight.BlockHeight
				fmt.Println("===== ", txHeight, nh)
				if nh-txHeight >= uint32(interval) {
					return true, txHeight, nil
				}
			}
		case <-nTimer.C:
			return false, 0, fmt.Errorf("neo tx confirmed timeout: %s", txHash)
		}
	}
}

func (n *Transaction) IsConfirmedOverHeightInterval(txHeight uint32, interval int64) bool {
	nHeight, err := n.Client().GetStateHeight()
	if err != nil {
		return false
	}
	nh := nHeight.BlockHeight
	return nh-txHeight > uint32(interval)
}

func (n *Transaction) ValidateAddress(addr string) error {
	return n.client.ValidateAddress(addr)
}

func (n *Transaction) RHashFromApplicationLog(hash string) (string, error) {
	if h, err := util.Uint256DecodeStringLE(hash); err == nil {
		if l, err := n.client.GetApplicationLog(h); err == nil {
			data, _ := json.MarshalIndent(l, "", "\t")
			fmt.Println(string(data))
			return "l", nil
		} else {
			return "", fmt.Errorf("get applicationLog: %s", err)
		}
	} else {
		return "", fmt.Errorf("decode string: %s", err)
	}
}
