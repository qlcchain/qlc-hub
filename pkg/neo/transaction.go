package neo

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
)

type Transaction struct {
	signer       *signer.SignerClient
	urls         []string
	url          string
	client       *client.Client
	contractLE   util.Uint160
	contractAddr string
	pendingTx    *sync.Map
	logger       *zap.SugaredLogger
}

func NewTransaction(urls []string, contractAddr string, signer *signer.SignerClient) (*Transaction, error) {
	c, err := client.New(context.Background(), urls[0], client.Options{})
	if err != nil {
		return nil, err
	}
	contract, err := util.Uint160DecodeStringLE(contractAddr)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		signer:       signer,
		urls:         urls,
		url:          urls[0],
		client:       c,
		contractLE:   contract,
		contractAddr: contractAddr,
		pendingTx:    new(sync.Map),
		logger:       log.NewLogger("neo/transaction"),
	}, nil
}

type TransactionParam struct {
	Params        []request.Param
	SignerAddress string
	Sysfee        util.Fixed8
	Netfee        util.Fixed8
	ROrigin       string
	RHash         string
	FuncName      string
	EmitIndex     string
}

func (n *Transaction) TxVerifyAndConfirmed(txHash string, interval int) error {
	hash, err := util.Uint256DecodeStringLE(hubUtil.RemoveHexPrefix(txHash))
	if err != nil {
		return fmt.Errorf("tx verify decode hash: %s, %s", err, txHash)
	}
	var txHeight uint32
	if txHeight, err = n.client.GetTransactionHeight(hash); err != nil {
		return fmt.Errorf("verify neo transaction: %s, tx:%s", err, hash)
	}
	if b, current := n.hasBlocksConfirmed(txHeight, int64(interval)); !b {
		return fmt.Errorf("confirmed neo transaction fail, %s,  %d/%d", txHash, txHeight, current)
	}
	if err := n.lockerEventFromApplicationLog(txHash); err != nil { // check if failed
		return err
	}
	return nil
}

func (n *Transaction) WaitTxVerifyAndConfirmed(txHash string, interval int) (uint32, error) {
	hash, err := util.Uint256DecodeStringLE(txHash)
	if err != nil {
		return 0, fmt.Errorf("tx verify decode hash: %s, %s", err, txHash)
	}
	client := n.Client()

	var txHeight uint32
	if txHeight, err = client.GetTransactionHeight(hash); err != nil {
		cTicker := time.NewTicker(3 * time.Second)
		cTimer := time.NewTimer(300 * time.Second) //tx on chain may need time
		for {
			select {
			case <-cTicker.C:
				txHeight, err = client.GetTransactionHeight(hash)
				if err != nil {
					n.logger.Debugf("get neo tx [%s] height err: %s", txHash, err)
				} else {
					goto HeightConfirmed
				}
			case <-cTimer.C:
				return 0, fmt.Errorf("tx(neo) by hash timeout: %s", txHash)
			}
		}
	}

HeightConfirmed:

	if b, _ := n.hasBlocksConfirmed(txHeight, int64(interval)); b {
		return txHeight, nil
	}
	nTicker := time.NewTicker(5 * time.Second)
	nTimer := time.NewTimer(time.Duration((interval+1)*121) * time.Second)
	for {
		select {
		case <-nTicker.C:
			if b, _ := n.hasBlocksConfirmed(txHeight, int64(interval)); b {
				if err := n.lockerEventFromApplicationLog(txHash); err != nil { // check if failed
					return 0, err
				} else {
					return txHeight, nil
				}
			}
		case <-nTimer.C:
			return 0, fmt.Errorf("neo tx confirmed timeout: %s", txHash)
		}
	}
}

func (n *Transaction) hasBlocksConfirmed(startHeight uint32, interval int64) (bool, uint32) {
	if interval < 0 {
		interval = 0
	}
	nHeight, err := n.client.GetStateHeight()
	if err != nil {
		return false, 0
	}
	nh := nHeight.BlockHeight
	n.logger.Debugf("current confirmed height (%d -> %d)", startHeight, nh)
	return nh-startHeight >= uint32(interval), nh
}

func (n *Transaction) ValidateAddress(addr string) error {
	return n.client.ValidateAddress(addr)
}

func (n *Transaction) SignData(address string, str string) (*proto.SignResponse, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return n.signer.Sign(proto.SignType_NEO, address, data)
}

func (n *Transaction) GetTransactionHeight(txHash string) (uint32, error) {
	hash, err := util.Uint256DecodeStringLE(txHash)
	if err != nil {
		return 0, fmt.Errorf("tx verify decode hash: %s", err)
	}
	return n.client.GetTransactionHeight(hash)
}

func (n *Transaction) Client() *client.Client {
	if err := n.client.Ping(); err == nil {
		return n.client
	} else {
		n.logger.Errorf("ping neo client: %s, %s ", err, n.urls[0])
		for _, url := range n.urls {
			if url != n.url {
				c, err := client.New(context.Background(), url, client.Options{})
				if err == nil {
					if err := c.Ping(); err == nil {
						n.client = c
						n.url = url
						return c
					} else {
						n.logger.Errorf("ping neo client: %s, %s", err, url)
					}
				} else {
					n.logger.Errorf("new neo client: %s, %s", err, url)
				}
			}
		}
	}
	return n.client
}

func (n *Transaction) ClientEndpoint() string {
	return n.url
}

func (n *Transaction) lockerEventFromApplicationLog(hash string) error {
	if h, err := util.Uint256DecodeStringLE(hash); err == nil {
		if l, err := n.client.GetApplicationLog(h); err == nil {
			//data, _ := json.MarshalIndent(l, "", "\t")
			//fmt.Println(string(data))
			for _, executions := range l.Executions {
				for _, events := range executions.Events {
					if events.Item.Type == smartcontract.ArrayType {
						return nil
					}
				}
			}
		} else {
			return fmt.Errorf("get applicationLog: %s, %s", err, hash)
		}
	}
	return fmt.Errorf("can not find lock event, txHash: %s", hash)
}
