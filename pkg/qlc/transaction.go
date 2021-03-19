package qlc

import (
	"sync"

	"github.com/nspcc-dev/neo-go/pkg/util"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
)

type Transaction struct {
	signer       *signer.SignerClient
	urls         []string
	url          string
	client       *qlcchain.QLCClient
	contractLE   util.Uint160
	contractAddr string
	pendingTx    *sync.Map
	logger       *zap.SugaredLogger
}

func NewTransaction(url string, signer *signer.SignerClient) (*Transaction, error) {
	c, err := qlcchain.NewQLCClient(url)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		signer: signer,
		client: c,
		logger: log.NewLogger("neo/transaction"),
	}, nil
}

func (n *Transaction) Client() *qlcchain.QLCClient {
	return n.client
}
