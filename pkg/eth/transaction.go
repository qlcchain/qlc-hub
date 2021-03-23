package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/signer"
)

type Transaction struct {
	signer     *signer.SignerClient
	urls       []string
	url        string
	client     *ethclient.Client
	contract   common.Address
	averageGas int64
	ctx        context.Context
	pendingTx  *sync.Map
	logger     *zap.SugaredLogger
}

func NewTransaction(urls []string, contract string) (*Transaction, error) {
	eClient, err := ethclient.Dial(urls[0])
	if err != nil || eClient == nil {
		return nil, fmt.Errorf("eth dail: %s", err)
	}
	return &Transaction{
		client:    eClient,
		urls:      urls,
		url:       urls[0],
		contract:  common.HexToAddress(contract),
		pendingTx: new(sync.Map),
		logger:    log.NewLogger("eth/transaction"),
	}, nil
}

func (t *Transaction) WaitTxVerifyAndConfirmed(txHash common.Hash, txHeight uint64, interval int64) error {
	cTicker := time.NewTicker(3 * time.Second)
	cTimer := time.NewTimer(600 * time.Second)
	client := t.Client()
	for {
		select {
		case <-cTicker.C:
			tx, p, err := client.TransactionByHash(context.Background(), txHash)
			if err != nil {
				t.logger.Errorf("eth tx by hash: %s , txHash: %s", err, txHash.String())
			}
			if tx != nil && !p { // if tx not found , p is false
				if txHeight == 0 {
					recepit, err := client.TransactionReceipt(context.Background(), txHash)
					if err == nil {
						txHeight = recepit.BlockNumber.Uint64()
						goto HeightConfirmed
					}
				} else {
					goto HeightConfirmed
				}
			}
		case <-cTimer.C:
			return fmt.Errorf("eth tx by hash timeout: %s", txHash)
		}
	}

HeightConfirmed:

	vTicker := time.NewTicker(5 * time.Second)
	vTimer := time.NewTimer(time.Duration((interval+1)*61) * time.Second)
	for {
		select {
		case <-vTicker.C:
			b, _ := t.HasConfirmedBlocksByHeight(int64(txHeight), interval)
			if b {
				return nil
			}
		case <-vTimer.C:
			return fmt.Errorf("confrimed eth tx by hash timeout: %s", txHash)
		}
	}
}

func (t *Transaction) GetBestBlockHeight() (int64, error) {
	block, err := t.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("BlockByNumber: %s", err)
	}
	return block.Number().Int64(), nil
}

func (t *Transaction) HasBlockConfirmed(txHash common.Hash, interval int64) (bool, error) {
	//tx, p, err := t.client.TransactionByHash(context.Background(), txHash)
	//if err != nil {
	//	return false, err
	//}
	//if tx != nil && !p { // if tx not found , p is false
	recepit, err := t.client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return false, fmt.Errorf("tx receipt: %s", err)
	}
	blockNumber := recepit.BlockNumber
	confirmed, _ := t.HasConfirmedBlocksByHeight(blockNumber.Int64(), interval)
	if !confirmed {
		return false, errors.New("block not confirmed")
	}
	return true, nil
	//}
	//return false, errors.New("tx not found")
}

func (t *Transaction) HasConfirmedBlocksByHeight(startHeight int64, interval int64) (bool, int64) {
	if interval < 0 {
		interval = 0
	}
	bestHeight, err := t.GetBestBlockHeight()
	if err != nil {
		return false, 0
	}
	return bestHeight-startHeight >= interval, bestHeight
}

func (t *Transaction) EthBalance(addr string) (int64, error) {
	address := common.HexToAddress(addr)
	balance, err := t.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return 0, err
	}
	return balance.Int64(), nil
}

func GetAccountByPriKey(priKey string) (*ecdsa.PrivateKey, common.Address, error) {
	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, errors.New("invaild public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, fromAddress, nil
}

func (t *Transaction) SyncBurnLog(txHash string) (*big.Int, common.Address, string, error) {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("TransactionReceipt, %s [%s]", err, txHash)
	}
	filterer, err := NewQLCChainFilterer(t.contract, t.client)
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("NewQLCChainFilterer, %s [%s]", err, txHash)
	}
	for _, log := range receipt.Logs {
		event, err := filterer.ParseBurn(*log)
		if err == nil && event != nil {
			return event.Amount, event.User, event.Nep5Addr, nil
		}
	}
	return nil, common.Address{}, "", fmt.Errorf("burn log not found, [%s]", txHash)
}

func (t *Transaction) SyncMintLog(txHash string) (*big.Int, common.Address, string, error) {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("TransactionReceipt, %s [%s]", err, txHash)
	}
	filterer, err := NewQLCChainFilterer(t.contract, t.client)
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("NewQLCChainFilterer, %s [%s]", err, txHash)
	}
	for _, log := range receipt.Logs {
		event, err := filterer.ParseMint(*log)
		if err == nil && event != nil {
			hash := common.BytesToHash(event.Nep5Hash[:])
			return event.Amount, event.User, hash.String(), nil
		}
	}
	return nil, common.Address{}, "", fmt.Errorf("burn log not found, [%s]", txHash)
}

func (t *Transaction) Client() *ethclient.Client {
	if _, err := t.client.BlockByNumber(context.Background(), nil); err == nil {
		return t.client
	} else {
		t.logger.Errorf("eth client: %s, %s ", err, t.url)
		for _, url := range t.urls {
			if url != t.url {
				c, err := ethclient.Dial(url)
				if err == nil && c != nil {
					if _, err := t.client.BlockByNumber(context.Background(), nil); err == nil {
						t.client = c
						t.url = url
						return c
					} else {
						t.logger.Errorf("eth client: %s, %s ", err, t.url)
					}
				} else {
					t.logger.Errorf("new eth client: %s, %s", err, url)
				}
			}
		}
	}
	return t.client
}

func (t *Transaction) ClientEndpoint() string {
	return t.url
}
