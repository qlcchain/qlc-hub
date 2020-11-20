package neo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
)

func (n *Transaction) UnsignedLockTransaction(userAddress, assetsAddr, rHash string, amount int) (string, string, error) {
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
	return n.CreateUnsignedTransaction(TransactionParam{
		Params:        params,
		SignerAddress: userAddress,
	})
}

func signatureVerify(unsignedData []byte, signature, publicKey, address string) ([]byte, []byte, error) {
	pk, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid publickey: %s", publicKey)
	}
	signBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid signature: %s", signature)
	}
	if pk.Address() != address {
		return nil, nil, fmt.Errorf("invaild publickey, publickey:%s, address:%s", publicKey, address)
	}
	hash := sha256.Sum256(unsignedData)
	if !pk.Verify(signBytes, hash[:]) {
		return nil, nil, fmt.Errorf("invaild signature, unsignedData:%s, signature:%s", hex.EncodeToString(unsignedData), signature)
	}
	return pk.GetVerificationScript(), signBytes, nil
}

func (n *Transaction) SendLockTransaction(txHash, signature, publicKey, address string) (string, error) {
	txObj, ok := n.pendingTx.Load(txHash)
	if !ok {
		return "", fmt.Errorf("tx not found: %s", txHash)
	}
	tx, ok := txObj.(*transaction.Transaction)
	if !ok {
		return "", fmt.Errorf("invalid tx : %s", txHash)
	}
	verificationScript, signatureBytes, err := signatureVerify(tx.GetSignedPart(), signature, publicKey, address)
	if err != nil {
		return "", err
	}
	tx.Scripts = append(tx.Scripts, transaction.Witness{
		InvocationScript:   append([]byte{byte(opcode.PUSHBYTES64)}, signatureBytes...),
		VerificationScript: verificationScript,
	})
	err = n.client.SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("failed sendning tx: %s", err)
	}
	return tx.Hash().StringLE(), nil
}
