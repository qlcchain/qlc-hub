package neo

import (
	"context"

	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

type NeoTransaction struct {
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
		client:   c,
		contract: contract,
		logger:   log.NewLogger("transaction"),
	}, nil
}

func (n *NeoTransaction) CreateTransaction(params []request.Param, wif string, sysfee util.Fixed8, netfee util.Fixed8) (string, error) {
	account, err := wallet.NewAccountFromWIF(wif)
	if err != nil {
		n.logger.Error(err)
		return "", err
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contract, params)
	if err != nil {
		n.logger.Error(err)
		return "", err
	}
	re, err := n.client.SignAndPushInvocationTx(scripts, account, sysfee, netfee)
	if err != nil {
		n.logger.Error(err)
		return "", err
	}
	return re.StringLE(), nil
}
