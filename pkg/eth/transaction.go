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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func GetTransactor(client *ethclient.Client, priKey, contract string) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := getTransactOpts(client, priKey)
	if err != nil {
		return nil, nil, err
	}
	instance, err := NewQLCChainTransactor(common.HexToAddress(contract), client)
	if err != nil {
		return nil, nil, fmt.Errorf("new transactor: %s", err)
	}
	return instance, auth, nil
}

func GetTransactorSession(client *ethclient.Client, priKey, contract string) (*QLCChainSession, error) {
	cAddress := common.HexToAddress(contract)
	c, err := NewQLCChain(cAddress, client)
	if err != nil {
		return nil, fmt.Errorf("new contract: %s", err)
	}
	auth, err := getTransactOpts(client, priKey)
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

func getTransactOpts(client *ethclient.Client, priKey string) (*bind.TransactOpts, error) {
	privateKey, fromAddress, err := GetAccountByPriKey(priKey)
	if err != nil {
		return nil, err
	}
	//todo rethink auth parameter settings
	ctx := context.Background()
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %s", err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
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

func GetHashTimer(client *ethclient.Client, contract string, rHash string) (*HashTimer, error) {
	instance, err := NewQLCChainCaller(common.HexToAddress(contract), client)
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

func TxVerifyAndConfirmed(txHash string, txHeight int64, interval int64, client *ethclient.Client) (bool, error) {
	cTicker := time.NewTicker(2 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			_, p, err := client.TransactionByHash(context.Background(), common.HexToHash(txHash))
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
			b, _ := HasConfirmedBlocksHeight(txHeight, interval, client)
			if b {
				return true, nil
			}
		case <-vTimer.C:
			return false, fmt.Errorf("eth tx by hash timeout: %s", txHash)
		}
	}
}

func GetBestBlockHeight(client *ethclient.Client) (int64, error) {
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("BlockByNumber: %s", err)
	}
	return block.Number().Int64(), nil
}

func HasConfirmedBlocksHeight(startHeight int64, interval int64, client *ethclient.Client) (bool, int64) {
	if interval < 0 {
		interval = 0
	}
	bestHeight, err := GetBestBlockHeight(client)
	if err != nil {
		return false, 0
	}
	return bestHeight-startHeight >= interval, bestHeight
}
