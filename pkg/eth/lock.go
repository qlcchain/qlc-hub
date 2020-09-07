package eth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

// withdraw
func (t *Transaction) UserLock(rHash, signerAddr, wrapperAddr string, amount int64) (string, error) {
	instance, opts, err := t.GetTransactor(signerAddr)
	if err != nil {
		return "", fmt.Errorf("UserLock/GetTransactor: %s", err)
	}

	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	bigAmount := big.NewInt(amount)
	tx, err := instance.DestoryLock(opts, rHashBytes, bigAmount, common.HexToAddress(wrapperAddr))
	if err != nil {
		return "", fmt.Errorf("UserLock/IssueUnlock: %s", err)
	}
	return tx.Hash().Hex(), nil
}

func (t *Transaction) WrapperUnlock(rHash, rOrigin, signerAddr string) (string, error) {
	instance, opts, err := t.GetTransactor(signerAddr)
	if err != nil {
		return "", fmt.Errorf("WrapperUnlock/GetTransactor: %s", err)
	}
	rOriginBytes, err := util.StringToBytes32(rOrigin)
	if err != nil {
		return "", err
	}
	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	tx, err := instance.DestoryUnlock(opts, rHashBytes, rOriginBytes)
	if err != nil {
		return "", fmt.Errorf("WrapperUnlock/DestoryUnlock: %s", err)
	}
	return tx.Hash().Hex(), nil
}

func (t *Transaction) UserFetch(rHash, signerAddr string) (string, error) {
	instance, opts, err := t.GetTransactor(signerAddr)
	if err != nil {
		return "", err
	}

	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	tx, err := instance.DestoryFetch(opts, rHashBytes)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// deposit
func (t *Transaction) WrapperLock(rHash, signerAddr string, amount int64) (string, error) {
	instance, opts, err := t.GetTransactor(signerAddr)
	if err != nil {
		return "", err
	}

	bigAmount := big.NewInt(amount)
	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	tx, err := instance.IssueLock(opts, rHashBytes, bigAmount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (t *Transaction) UserUnlock(rHash, rOrigin, userPriKey string) (string, error) {
	instance, opts, err := t.GetTransactor(userPriKey)
	if err != nil {
		return "", fmt.Errorf("UserUnlock/GetTransactor: %s", err)
	}

	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	rOriginBytes, err := util.StringToBytes32(rOrigin)
	if err != nil {
		return "", err
	}
	tx, err := instance.IssueUnlock(opts, rHashBytes, rOriginBytes)
	if err != nil {
		return "", fmt.Errorf("UserUnlock/IssueUnlock: %s", err)
	}
	return tx.Hash().Hex(), nil
}

func (t *Transaction) WrapperFetch(rHash, signerAddr string) (string, error) {
	instance, opts, err := t.GetTransactor(signerAddr)
	if err != nil {
		return "", err
	}

	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		return "", err
	}
	tx, err := instance.IssueFetch(opts, rHashBytes)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}
