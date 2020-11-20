package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
)

type Transaction struct {
	signer     *signer.SignerClient
	client     *ethclient.Client
	contract   common.Address
	averageGas int64
	ctx        context.Context
	logger     *zap.SugaredLogger
}

func NewTransaction(client *ethclient.Client, contract string) *Transaction {
	return &Transaction{
		client:   client,
		contract: common.HexToAddress(contract),
		logger:   log.NewLogger("eth/transaction"),
	}
}

func (t *Transaction) TxVerifyAndConfirmed(txHash common.Hash, txHeight uint64, interval int64) error {
	cTicker := time.NewTicker(5 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			tx, p, err := t.client.TransactionByHash(context.Background(), txHash)
			if err != nil {
				t.logger.Errorf("eth tx by hash: %s , txHash: %s", err, txHash.String())
			}
			if tx != nil && !p { // if tx not found , p is false
				goto HeightConfirmed
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
			b, _ := t.HasConfirmedBlocksHeight(int64(txHeight), interval)
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

func (t *Transaction) HasConfirmedBlocksHeight(startHeight int64, interval int64) (bool, int64) {
	if interval < 0 {
		interval = 0
	}
	bestHeight, err := t.GetBestBlockHeight()
	if err != nil {
		return false, 0
	}
	return bestHeight-startHeight > interval, bestHeight
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

func (t *Transaction) Client() *ethclient.Client {
	return t.client
}

func (t *Transaction) SyncLog(txHash string) (int64, error) {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return 0, fmt.Errorf("1, %s", err)
	}
	filterer, err := NewQLCChainFilterer(t.contract, t.client)
	if err != nil {
		return 0, fmt.Errorf("11, %s", err)
	}
	for _, log := range receipt.Logs {
		event, err := filterer.ParseMint(*log)

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(event.Amount.String(), event.User.String(), hex.EncodeToString(event.Nep5Hash[:]))
		}
	}
	return 0, nil
}
