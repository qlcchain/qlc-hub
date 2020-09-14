package neo

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
)

// deposit
func (n *Transaction) UserLock(userAddress, assetsAddr, rHash string, amount int) (string, error) {
	params := []request.Param{
		FunctionName("userLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(userAddress),
			IntegerTypeParam(amount),
			AddressParam(assetsAddr),
			IntegerTypeParam(10), //todo
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:        params,
		SignerAddress: userAddress,
	})
	if err != nil {
		return "", fmt.Errorf("UserLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) WrapperUnlock(rOrigin, signerAddress, userEthAddress string) (string, error) {
	params := []request.Param{
		FunctionName("wrapperUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			ArrayTypeParam(userEthAddress),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddress,
		ROrigin:       rOrigin,
		FuncName:      "wrapperUnlock",
		EmitIndex:     "1",
	})
	if err != nil {
		return "", fmt.Errorf("wrapperUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundUser(rOrigin string, signerAddress string) (string, error) {
	params := []request.Param{
		FunctionName("refundUser"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddress,
		ROrigin:       rOrigin,
		FuncName:      "refundUser",
		EmitIndex:     "2",
	})
	if err != nil {
		return "", fmt.Errorf("refundUser/createTransaction: %s", err)
	}
	return r, nil
}

// withdraw

func (n *Transaction) WrapperLock(assetsAddr, userEthAddress, rHash string, amount, timerInterval int) (string, error) {
	params := []request.Param{
		FunctionName("wrapperLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(assetsAddr),
			IntegerTypeParam(amount),
			ArrayTypeParam(userEthAddress),
			IntegerTypeParam(timerInterval),
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:        params,
		SignerAddress: assetsAddr,
	})
	if err != nil {
		return "", fmt.Errorf("wrapperLock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) UserUnlock(rOrigin, userAddr, signerAddress string) (string, error) {
	params := []request.Param{
		FunctionName("userUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(userAddr),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddress,
		ROrigin:       rOrigin,
		FuncName:      "userUnlock",
		EmitIndex:     "1",
	})
	if err != nil {
		return "", fmt.Errorf("userUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundWrapper(rHash, signerAddr string) (string, error) {
	params := []request.Param{
		FunctionName("refundWrapper"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddr,
		FuncName:      "refundWrapper",
		EmitIndex:     "1",
	})
	if err != nil {
		return "", fmt.Errorf("refundWrapper/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) DeleteSwapInfo(rHash, signerAddr string) (string, error) {
	params := []request.Param{
		FunctionName("deleteSwapInfo"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddr,
		FuncName:      "deleteSwapInfo",
		EmitIndex:     "1",
	})
	if err != nil {
		return "", fmt.Errorf("deleteSwapInfo/createTransaction: %s", err)
	}
	return r, nil
}
