package eth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/qlcchain/qlc-hub/pkg/util"
)

// withdraw
func UserLock(rHash, userPriKey, wrapperAddr, contract string, amount int64, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, userPriKey, contract)
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

func WrapperUnlock(rHash, rOrigin, wrapperPrikey, contract string, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, wrapperPrikey, contract)
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

func UserFetch(rHash, userPrikey, contract string, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, userPrikey, contract)
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
func WrapperLock(rHash, wrapperPrikey, contract string, amount int64, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, wrapperPrikey, contract)
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

func UserUnlock(rHash, rOrigin, userPriKey, contract string, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, userPriKey, contract)
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

func WrapperFetch(rHash, wrapperPrikey, contract string, client *ethclient.Client) (string, error) {
	instance, opts, err := GetTransactor(client, wrapperPrikey, contract)
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
