package neo

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/pkg/log"
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
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	// add witness
	script := io.NewBufBinWriter()
	emit.String(script.BinWriter, param.ROrigin)
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, param.FuncName)
	tx.Scripts = append(tx.Scripts, transaction.Witness{
		InvocationScript:   script.Bytes(),
		VerificationScript: []byte{},
	})

	if err := account.SignTx(tx); err != nil {
		return "", fmt.Errorf("signTx: %s", err)
	}

	if err := n.client.SendRawTransaction(tx); err != nil {
		return "", fmt.Errorf("sendRawTransaction: %s", err)
	}
	n.logger.Debugf("transaction successfully: %s", tx.Hash().StringLE())
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

type SwapInfo struct {
	rHash   string
	rOrigin string
	state   string
	amount  *big.Int
}

func (n *Transaction) QuerySwapInfo(rHash string) (*SwapInfo, error) {
	_, err := n.querySwapInfo(rHash)
	if err != nil {
		return nil, err
	}

	// convert result to SwapInfo
	return &SwapInfo{}, nil
}

func (n *Transaction) querySwapInfo(rHash string) (*result.Invoke, error) {
	//hash, err := util.Uint160DecodeStringLE(rHash)
	params := []smartcontract.Parameter{
		{
			Type:  smartcontract.StringType,
			Value: rHash,
		},
	}
	r, err := n.client.InvokeFunction(n.contractAddr, "querySwapInfo", params, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func TxVerifyAndConfirmed(txHash string, interval int, c *Transaction) (bool, uint32, error) {
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
			txHeight, err = c.client.GetTransactionHeight(hash)
			if err != nil {
				fmt.Println("======= ", err)
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
			nHeight, err := c.Client().GetStateHeight()
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
