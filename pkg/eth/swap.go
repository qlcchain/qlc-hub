package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func (t *Transaction) getTransactor(signerAccount string) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := t.getTransactOpts(signerAccount)
	if err != nil {
		return nil, nil, err
	}
	instance, err := NewQLCChainTransactor(t.contract, t.client)
	if err != nil {
		return nil, nil, fmt.Errorf("new transactor: %s", err)
	}
	return instance, auth, nil
}

func (t *Transaction) getTransactOpts(signerAddr string) (*bind.TransactOpts, error) {
	privateKey, fromAddress, err := GetAccountByPriKey(signerAddr)
	if err != nil {
		return nil, err
	}
	//todo rethink auth parameter settings
	nonce, err := t.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}

	suggestPrice, err := t.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("best gas price: %s", err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(8000000) // in units
	auth.GasPrice = suggestPrice
	return auth, nil
}

func (t *Transaction) Mint(signerAccount string, amount *big.Int, neoHash string, signature string) (string, error) {
	instance, opts, err := t.getTransactor(signerAccount)
	if err != nil {
		return "", err
	}
	nHashBytes, err := util.HexStringToBytes32(neoHash)
	if err != nil {
		return "", err
	}
	signBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}
	tx, err := instance.Mint(opts, amount, nHashBytes, signBytes)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (t *Transaction) Burn(signerAccount string, amount *big.Int, neoHash string, signature string) (string, error) {
	instance, opts, err := t.getTransactor(signerAccount)
	if err != nil {
		return "", err
	}
	nHashBytes, err := util.HexStringToBytes32(neoHash)
	if err != nil {
		return "", err
	}
	signBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}
	tx, err := instance.Mint(opts, amount, nHashBytes, signBytes)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}
