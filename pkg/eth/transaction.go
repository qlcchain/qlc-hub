package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetTransactor(client *ethclient.Client, priKey, contract string) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := getTransactOpts(client, priKey)
	if err != nil {
		return nil, nil, err
	}
	address := common.HexToAddress(contract)
	instance, err := NewQLCChainTransactor(address, client)
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
	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invaild public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

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

type State int

const (
	IssueLock State = iota
	IssueUnlock
	IssueFetch
	DestroyLock
	DestroyUnlock
	DestroyFetch
)
