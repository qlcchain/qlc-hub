package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func (t *Transaction) getQGasTransactor(signerAccount string) (transactor *QGasChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := t.getQGasTransactOpts(signerAccount)
	if err != nil {
		return nil, nil, err
	}
	instance, err := NewQGasChainTransactor(t.contract, t.client)
	if err != nil {
		return nil, nil, fmt.Errorf("new transactor: %s", err)
	}
	return instance, auth, nil
}

func (t *Transaction) getQGasTransactOpts(signerAddr string) (*bind.TransactOpts, error) {
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
	auth.GasPrice = big.NewInt(0).Mul(suggestPrice, big.NewInt(10))
	return auth, nil
}

func (t *Transaction) QGasMint(signerAccount string, amount *big.Int, neoHash string, signature string) (string, error) {
	instance, opts, err := t.getQGasTransactor(signerAccount)
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

func (t *Transaction) QGasBurn(signerAccount string, nep5Addr string, amount *big.Int) (string, error) {
	instance, opts, err := t.getQGasTransactor(signerAccount)
	if err != nil {
		return "", err
	}
	tx, err := instance.Burn(opts, nep5Addr, amount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// if key not found, return is 0
func (t *Transaction) GetQGasLockedAmountByQLCTxHash(qlcHash string) (*big.Int, error) {
	instance, err := NewQGasChainCaller(t.contract, t.client)
	if err != nil {
		return nil, err
	}
	nHashBytes, err := util.HexStringToBytes32(util.RemoveHexPrefix(qlcHash))
	if err != nil {
		return nil, err
	}
	return instance.LockedAmount(&bind.CallOpts{}, nHashBytes)
}

func (t *Transaction) QGasSyncBurnLog(txHash string) (*big.Int, common.Address, string, error) {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("TransactionReceipt, %s [%s]", err, txHash)
	}
	filterer, err := NewQGasChainFilterer(t.contract, t.client)
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("NewQLCChainFilterer, %s [%s]", err, txHash)
	}
	for _, log := range receipt.Logs {
		event, err := filterer.ParseBurn(*log)
		if err == nil && event != nil {
			return event.Amount, event.User, event.QlcAddr, nil
		}
	}
	return nil, common.Address{}, "", fmt.Errorf("burn log not found, [%s]", txHash)
}

func (t *Transaction) QGasSyncMintLog(txHash string) (*big.Int, common.Address, string, error) {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("TransactionReceipt, %s [%s]", err, txHash)
	}
	filterer, err := NewQGasChainFilterer(t.contract, t.client)
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("NewQLCChainFilterer, %s [%s]", err, txHash)
	}
	for _, log := range receipt.Logs {
		event, err := filterer.ParseMint(*log)
		if err == nil && event != nil {
			qlcHash, err := qlctypes.BytesToHash(event.QlcHash[:])
			if err != nil {
				return nil, common.Address{}, "", fmt.Errorf("invalid hash, %s", err)
			}
			return event.Amount, event.User, qlcHash.String(), nil
		}
	}
	return nil, common.Address{}, "", fmt.Errorf("burn log not found, [%s]", txHash)
}
