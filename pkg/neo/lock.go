package neo

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"

	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
)

// deposit
func (n *Transaction) UserLock(publicKey, wrapperAddress, rHash string, amount int) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("userLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(pubKey.Address()),
			IntegerTypeParam(amount),
			AddressParam(wrapperAddress),
			IntegerTypeParam(10), //todo
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
	})
	if err != nil {
		return "", fmt.Errorf("UserLock/CreateTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) WrapperUnlock(rOrigin, publicKey, userEthAddress string) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("wrapperUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(pubKey.Address()),
			ArrayTypeParam(userEthAddress),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
		ROrigin:   rOrigin,
		FuncName:  "wrapperUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("wrapperUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundUser(rOrigin string, publicKey string) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", fmt.Errorf("new account from wif: %s", err)
	}
	params := []request.Param{
		FunctionName("refundUser"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(pubKey.Address()),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
		ROrigin:   rOrigin,
		FuncName:  "refundUser",
	})
	if err != nil {
		return "", fmt.Errorf("refundUser/createTransaction: %s", err)
	}
	return r, nil
}

// withdraw

func (n *Transaction) WrapperLock(publicKey, userEthAddress, rHash string, amount int) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("wrapperLock"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(pubKey.Address()),
			IntegerTypeParam(amount),
			ArrayTypeParam(userEthAddress),
			IntegerTypeParam(10),
		}),
	}
	r, err := n.CreateTransaction(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
	})
	if err != nil {
		return "", fmt.Errorf("wrapperLock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) UserUnlock(rOrigin, publicKey string) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", err
	}
	params := []request.Param{
		FunctionName("userUnlock"),
		ArrayParams([]request.Param{
			StringTypeParam(rOrigin),
			AddressParam(pubKey.Address()),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
		ROrigin:   rOrigin,
		FuncName:  "userUnlock",
	})
	if err != nil {
		return "", fmt.Errorf("userUnlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) RefundWrapper(rHash, publicKey string) (string, error) {
	pubKey, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return "", err
	}

	params := []request.Param{
		FunctionName("refundWrapper"),
		ArrayParams([]request.Param{
			ArrayTypeParam(rHash),
			AddressParam(pubKey.Address()),
		}),
	}
	r, err := n.CreateTransactionAppendWitness(TransactionParam{
		Params:    params,
		PublicKey: pubKey,
		RHash:     rHash,
		FuncName:  "refundWrapper",
	})
	if err != nil {
		return "", fmt.Errorf("refundWrapper/createTransaction: %s", err)
	}
	return r, nil
}
