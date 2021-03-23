package qlc

import (
	"fmt"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/util"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/signer"
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
		logger: log.NewLogger("qlc/transaction"),
	}, nil
}

func (t *Transaction) Client() *qlcchain.QLCClient {
	return t.client
}

func (t *Transaction) WaitBlockConfirmedOnChain(hash qlctypes.Hash) bool {
	ticker := time.NewTicker(1 * time.Second)
	for i := 0; i < 100; i++ {
		<-ticker.C
		b := t.CheckBlockOnChain(hash)
		if b {
			t.logger.Infof("check block (%s) confirmed status true", hash.String())
			return true
		}
	}
	return false
}

func (t *Transaction) CheckBlockOnChain(hash qlctypes.Hash) bool {
	b, err := t.Client().Ledger.BlockConfirmedStatus(hash)
	if !b || err != nil {
		return false
	}
	return true
}

func (t *Transaction) ProcessAndWaitConfirmed(block *qlctypes.StateBlock) error {
	h, err := t.Client().Ledger.Process(block)
	if err != nil {
		return fmt.Errorf("process block: %s", err)
	}
	if t.WaitBlockConfirmedOnChain(h) {
		return nil
	} else {
		return fmt.Errorf("verify block: %s", err)
	}
}

func (t *Transaction) GetSwapInfoHashByWithdrawSendBlock(hash qlctypes.Hash, store *gorm.DB) (*types.QGasSwapInfo, error) {
	sendBlk, err := t.Client().Ledger.BlockInfo(hash)
	if err != nil {
		return nil, fmt.Errorf("QGas can not get block: %s", err)
	}
	swapParam, err := t.Client().QGasSwap.ParseWithdrawParam(sendBlk.Data)
	if err != nil {
		return nil, fmt.Errorf("QGas parse withdraw param: %s", err)
	}
	swapInfo, err := db.GetQGasSwapInfoByUniqueID(store, swapParam.LinkHash.String(), types.ETH)
	if err != nil {
		return nil, fmt.Errorf("QGas pledge withdraw not found: %s", err)
	}
	return swapInfo, nil
}
