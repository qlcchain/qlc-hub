package neo

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

type NeoTransaction struct {
	url      string
	client   *client.Client
	contract util.Uint160
	logger   *zap.SugaredLogger
}

func NewNeoTransaction(url, contractAddr string) (*NeoTransaction, error) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		return nil, err
	}
	contract, err := util.Uint160DecodeStringLE(contractAddr)
	if err != nil {
		return nil, err
	}
	return &NeoTransaction{
		url:      url,
		client:   c,
		contract: contract,
		logger:   log.NewLogger("neo/transaction"),
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

func (n *NeoTransaction) CreateTransaction(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("NewAccountFromWIF: %s", err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contract, param.Params)
	if err != nil {
		return "", fmt.Errorf("CreateFunctionInvocationScript: %s", err)
	}
	re, err := n.client.SignAndPushInvocationTx(scripts, account, param.Sysfee, param.Netfee)
	if err != nil {
		return "", fmt.Errorf("SignAndPushInvocationTx: %s", err)
	}
	n.logger.Debugf("transaction successfully: %s", re.StringLE())
	return re.StringLE(), nil
}

func (n *NeoTransaction) CreateTransactionAppendWitness(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("new account: %s", err)
	}
	accountUint, err := address.StringToUint160(account.Address)
	if err != nil {
		return "", fmt.Errorf("address uint160: %s", err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contract, param.Params)
	if err != nil {
		return "", fmt.Errorf("create script: %s", err)
	}
	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(accountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  n.contract.BytesBE(),
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

func (n *NeoTransaction) Client() *client.Client {
	return n.client
}
