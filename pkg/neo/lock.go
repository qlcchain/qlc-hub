package neo

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

func UserLock(userWif, wrapperAddress, rHash string, amount int, c *Transaction) (string, error) {
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
	r, err := c.CreateTransaction(TransactionParam{
		Params: params,
		Wif:    userWif,
	})
	if err != nil {
		return "", fmt.Errorf("UserLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func WrapperUnlock(rOrigin, wrapperWif, userEthAddress string, c *Transaction) (string, error) {
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
	r, err := c.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		ROrigin:  rOrigin,
		FuncName: "wrapperUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("WrapperUnlock/CreateTransaction: %s", err)
	}
	return r, nil
}

// withdraw

func WrapperLock(wrapperWif, userEthAddress, rHash string, amount int, c *Transaction) (string, error) { //todo set int64
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
	r, err := c.CreateTransaction(TransactionParam{
		Params: params,
		Wif:    wrapperWif,
	})
	if err != nil {
		return "", fmt.Errorf("WrapperLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func UserUnlock(rOrigin, userWif string, c *Transaction) (string, error) {
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
	r, err := c.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      userWif,
		ROrigin:  rOrigin,
		FuncName: "userUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("UserUnlock/CreateTransaction: %s", err)
	}
	return r, nil
}

func RefundWrapper(rOrigin, wrapperWif string, c *Transaction) (string, error) {
	wrapperAccount, err := wallet.NewAccountFromWIF(wrapperWif)
	if err != nil {
		return "", err
	}

	params := []request.Param{
		FunctionName("refundWrapper"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(wrapperAccount.Address),
		}),
	}
	r, err := c.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Wif:      wrapperWif,
		ROrigin:  rOrigin,
		FuncName: "refundWrapper",
	})
	if err != nil {
		return "", fmt.Errorf("UserUnlock/CreateTransaction: %s", err)
	}
	return r, nil
}
