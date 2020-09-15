package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
)

type Transaction struct {
	signer     *signer.SignerClient
	client     *ethclient.Client
	contract   string
	averageGas int64
	ctx        context.Context
	gasUrl     string
	logger     *zap.SugaredLogger
}

func NewTransaction(client *ethclient.Client, signer *signer.SignerClient, ctx context.Context, gasUrl string, contract string) *Transaction {
	t := &Transaction{
		signer:   signer,
		client:   client,
		contract: contract,
		ctx:      ctx,
		gasUrl:   gasUrl,
		logger:   log.NewLogger("eth/transaction"),
	}
	go t.updateAverageGas()
	return t
}

func (t *Transaction) GetTransactor(signerAddr string) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := t.getTransactOpts(signerAddr)
	if err != nil {
		return nil, nil, err
	}
	instance, err := NewQLCChainTransactor(common.HexToAddress(t.contract), t.client)
	if err != nil {
		return nil, nil, fmt.Errorf("new transactor: %s", err)
	}
	return instance, auth, nil
}

func (t *Transaction) GetTransactorSession(ethAddress string) (*QLCChainSession, error) {
	cAddress := common.HexToAddress(t.contract)
	c, err := NewQLCChain(cAddress, t.client)
	if err != nil {
		return nil, fmt.Errorf("new contract: %s", err)
	}
	auth, err := t.getTransactOpts(ethAddress)
	if err != nil {
		return nil, err
	}
	// Pre-configured contract sessions
	session := &QLCChainSession{
		Contract: c,
		CallOpts: bind.CallOpts{
			Pending: true, //Whether to operate on the pending state or the last known one
		},
		TransactOpts: *auth,
	}
	return session, nil
}

func (t *Transaction) getTransactOpts(signerAddr string) (*bind.TransactOpts, error) {
	addr := common.HexToAddress(signerAddr)
	//todo rethink auth parameter settings
	nonce, err := t.client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}

	gasPrice, err := t.GetBestGas()
	if err != nil {
		return nil, fmt.Errorf("best gas price: %s", err)
	}
	auth := &bind.TransactOpts{
		From: addr,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != addr {
				t.logger.Error("no authorize account")
				return nil, errors.New("not authorized to sign this account")
			}
			signature, err := t.signer.Sign(proto.SignType_ETH, signerAddr, signer.Hash(tx).Bytes())
			if err != nil {
				t.logger.Error(err)
				return nil, err
			}
			if sign, err := tx.WithSignature(signer, signature.Sign); err != nil {
				t.logger.Error(err)
				return nil, err
			} else {
				return sign, nil
			}
		},
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(8000000) // in units
	auth.GasPrice = gasPrice
	t.logger.Debugf("eth tx auth, gasLimit: %d, gasPrice: %d, nonce: %d ", auth.GasLimit, gasPrice, nonce)
	return auth, nil
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

type State int

const (
	IssueLock State = iota
	IssueUnlock
	IssueFetch
	DestroyLock
	DestroyUnlock
	DestroyFetch
)

func StateValueToString(value int) string {
	switch State(value) {
	case IssueLock:
		return "IssueLock"
	case IssueUnlock:
		return "IssueUnLock"
	case IssueFetch:
		return "IssueFetch"
	case DestroyLock:
		return "DestroyLock"
	case DestroyUnlock:
		return "DestroyUnlock"
	case DestroyFetch:
		return "DestroyFetch"
	default:
		return "Invalid"
	}
}

func (t *Transaction) GetHashTimer(rHash string) (*HashTimer, error) {
	instance, err := NewQLCChainCaller(common.HexToAddress(t.contract), t.client)
	if err != nil {
		return nil, fmt.Errorf("eth/NewQLCChainCaller: %s", err)
	}
	rHashByte32, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return nil, fmt.Errorf("eth/HexStringToBytes32: %s", err)
	}
	var callops = bind.CallOpts{}
	rOrigin, amount, addr, lockno, unlockno, err := instance.HashTimer(&callops, rHashByte32)
	if err != nil {
		return nil, fmt.Errorf("get hashTimer: %s", err)
	}
	return &HashTimer{
		RHash:          rHash,
		Amount:         amount,
		ROrigin:        string(rOrigin[:]),
		UserAddr:       util.RemoveHexPrefix(addr.String()),
		LockedHeight:   uint32(lockno.Int64()),
		UnlockedHeight: uint32(unlockno.Int64()),
	}, nil
}

type HashTimer struct {
	RHash          string   `json:"rHash"`
	ROrigin        string   `json:"rOrigin"`
	Amount         *big.Int `json:"amount"`
	UserAddr       string   `json:"userAddr"`
	LockedHeight   uint32   `json:"lockedHeight"`
	UnlockedHeight uint32   `json:"unlockedHeight"`
}

func (h *HashTimer) String() string {
	v, _ := json.Marshal(h)
	return string(v)
}

func (t *Transaction) TxVerifyAndConfirmed(txHash string, txHeight int64, interval int64) error {
	cTicker := time.NewTicker(6 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			tx, p, err := t.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
			if err != nil {
				t.logger.Debugf("eth tx by hash: %s , txHash: %s", err, txHash)
			}
			if tx != nil && !p { // if tx not found , p is false
				goto HeightConfirmed
			}
		case <-cTimer.C:
			return fmt.Errorf("eth tx by hash timeout: %s", txHash)
		}
	}

HeightConfirmed:

	vTicker := time.NewTicker(6 * time.Second)
	vTimer := time.NewTimer(time.Duration((interval+1)*61) * time.Second)
	for {
		select {
		case <-vTicker.C:
			b, _ := t.HasConfirmedBlocksHeight(txHeight, interval)
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

func (t *Transaction) Client() *ethclient.Client {
	return t.client
}

func (t *Transaction) Balance(addr string) (int64, error) {
	address := common.HexToAddress(addr)
	balance, err := t.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return 0, err
	}
	return balance.Int64(), nil
}
