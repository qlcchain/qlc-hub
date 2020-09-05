package neo

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
)

// deposit
func (n *Transaction) UserLock(address, wrapperAddress, rHash string, amount int) (string, error) {
	params := []request.Param{
		FunctionName("userLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(address),
			IntegerTypeParam(amount),
			AddressParam(wrapperAddress),
			IntegerTypeParam(10), //todo
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:  params,
		Address: address,
	})
	if err != nil {
		return "", fmt.Errorf("UserLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) WrapperUnlock(rOrigin, address, userEthAddress string) (string, error) {
	params := []request.Param{
		FunctionName("wrapperUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(address),
			ArrayTypeParam(userEthAddress),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Address:  address,
		ROrigin:  rOrigin,
		FuncName: "wrapperUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("wrapperUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundUser(rOrigin string, address string) (string, error) {
	params := []request.Param{
		FunctionName("refundUser"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(address),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Address:  address,
		ROrigin:  rOrigin,
		FuncName: "refundUser",
	})
	if err != nil {
		return "", fmt.Errorf("refundUser/createTransaction: %s", err)
	}
	return r, nil
}

// withdraw

func (n *Transaction) WrapperLock(address, userEthAddress, rHash string, amount int) (string, error) {
	params := []request.Param{
		FunctionName("wrapperLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(address),
			IntegerTypeParam(amount),
			ArrayTypeParam(userEthAddress),
			IntegerTypeParam(10),
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:  params,
		Address: address,
	})
	if err != nil {
		return "", fmt.Errorf("wrapperLock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) UserUnlock(rOrigin, address string) (string, error) {
	params := []request.Param{
		FunctionName("userUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(address),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Address:  address,
		ROrigin:  rOrigin,
		FuncName: "userUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("userUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundWrapper(rHash, address string) (string, error) {
	params := []request.Param{
		FunctionName("refundWrapper"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(address),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:   params,
		Address:  address,
		RHash:    rHash,
		FuncName: "refundWrapper",
	})
	if err != nil {
		return "", fmt.Errorf("refundWrapper/createTransaction: %s", err)
	}
	return r, nil
}
