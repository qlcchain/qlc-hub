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
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
)

type Transaction struct {
	signer   *signer.SignerClient
	client   *ethclient.Client
	contract string
	logger   *zap.SugaredLogger
}

func NewTransaction(client *ethclient.Client, signer *signer.SignerClient, contract string) *Transaction {
	return &Transaction{
		signer:   signer,
		client:   client,
		contract: contract,
		logger:   log.NewLogger("eth/transaction"),
	}
}

func (t *Transaction) GetTransactor(ethAddress string) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	//TODO: fix it
	auth, err := t.getTransactOpts(ethAddress)
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
	//TODO: fix it
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

func (t *Transaction) getTransactOpts(ethAddr string) (*bind.TransactOpts, error) {
	addr := common.HexToAddress(ethAddr)
	//todo rethink auth parameter settings
	ctx := context.Background()
	nonce, err := t.client.PendingNonceAt(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}

	gasPrice, err := t.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %s", err)
	}
	auth := &bind.TransactOpts{
		From: addr,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != addr {
				t.logger.Error("no authorize account")
				return nil, errors.New("not authorized to sign this account")
			}
			signature, err := t.signer.Sign(proto.SignType_ETH, ethAddr, signer.Hash(tx).Bytes())
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

func (t *Transaction) TxVerifyAndConfirmed(txHash string, txHeight int64, interval int64) (bool, error) {
	cTicker := time.NewTicker(2 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			_, p, err := t.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
			if err != nil {
				return false, fmt.Errorf("eth tx by hash: %s", err)
			}
			if !p {
				goto HeightConfirmed
			}
		case <-cTimer.C:
			return false, fmt.Errorf("eth tx by hash timeout: %s", txHash)
		}
	}

HeightConfirmed:

	vTicker := time.NewTicker(5 * time.Second)
	vTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-vTicker.C:
			b, _ := t.HasConfirmedBlocksHeight(txHeight, interval)
			if b {
				return true, nil
			}
		case <-vTimer.C:
			return false, fmt.Errorf("eth tx by hash timeout: %s", txHash)
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
	return bestHeight-startHeight >= interval, bestHeight
}

func (t *Transaction) Client() *ethclient.Client {
	return t.client
}
