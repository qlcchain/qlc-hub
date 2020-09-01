package neo

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

func (n *Transaction) UserLock(userWif, wrapperAddress, rHash string, amount int) (string, error) {
	userAccount, err := wallet.NewAccountFromWIF(userWif)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("userLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(userAccount.Address),
			IntegerTypeParam(amount),
			AddressParam(wrapperAddress),
			IntegerTypeParam(10), //todo
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params: params,
		Wif:    userWif,
	})
	if err != nil {
		return "", fmt.Errorf("UserLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) WrapperUnlock(rOrigin, wrapperWif, userEthAddress string) (string, error) {
	wrapperAccount, err := wallet.NewAccountFromWIF(wrapperWif)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("wrapperUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(wrapperAccount.Address),
			ArrayTypeParam(userEthAddress),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		ROrigin:  rOrigin,
		FuncName: "wrapperUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("wrapperUnlock/createTransaction: %s", err)
	}
	return r, nil
}

// withdraw

func (n *Transaction) WrapperLock(wrapperWif, userEthAddress, rHash string, amount int) (string, error) { //todo set int64
	wrapperAccount, err := wallet.NewAccountFromWIF(wrapperWif)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("wrapperLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(wrapperAccount.Address),
			IntegerTypeParam(amount),
			ArrayTypeParam(userEthAddress),
			IntegerTypeParam(10), //todo setting
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params: params,
		Wif:    wrapperWif,
	})
	if err != nil {
		return "", fmt.Errorf("wrapperLock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) UserUnlock(rOrigin, userWif string) (string, error) {
	userAccount, err := wallet.NewAccountFromWIF(userWif)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("userUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(userAccount.Address),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      userWif,
		ROrigin:  rOrigin,
		FuncName: "userUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("userUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) WrapperFetch(rHash, wrapperWif string) (string, error) {
	wrapperAccount, err := wallet.NewAccountFromWIF(wrapperWif)
	if err != nil {
		return "", err
	}

	params := []request.Param{
		FunctionName("refundWrapper"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(wrapperAccount.Address),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		RHash:    rHash,
		FuncName: "refundWrapper",
	})
	if err != nil {
		return "", fmt.Errorf("wrapperFetch/createTransaction: %s", err)
	}
	return r, nil
}
