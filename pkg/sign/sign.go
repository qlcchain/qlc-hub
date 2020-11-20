package sign

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client *ethclient.Client
}

func NewClient(url string) *Client {
	c, err := ethclient.Dial("")
	if err != nil {
		return nil
	}
	return &Client{
		client: c,
	}
}

func (c *Client) GetQLCChainTransactor(priKey string, address common.Address) (transactor *QLCChainTransactor, opts *bind.TransactOpts, err error) {
	auth, err := c.getTransactOpts(priKey)
	if err != nil {
		return nil, nil, err
	}
	instance, err := NewQLCChainTransactor(address, c.client)
	if err != nil {
		return nil, nil, fmt.Errorf("new transactor: %s", err)
	}
	return instance, auth, nil
}

func (c *Client) getTransactOpts(priKey string) (*bind.TransactOpts, error) {
	privateKey, fromAddress, err := GetAccountByPriKey(priKey)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %s", err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(8000000) // in units
	auth.GasPrice = gasPrice
	return auth, nil
}

func GetAccountByPriKey(priKey string) (*ecdsa.PrivateKey, common.Address, error) {
	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("crypto hex to ecdsa: %s", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, errors.New("invaild public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, fromAddress, nil
}

func packed(amount *big.Int, userAddr string, nHash string) ([]byte, error) {
	packedBytes := make([]byte, 0)

	amountBytes := bytes.Repeat([]byte{0}, 32)
	aBytes := amount.Bytes()
	copy(amountBytes[len(amountBytes)-len(aBytes):], aBytes)
	packedBytes = append(packedBytes, amountBytes...)

	addr := common.HexToAddress(userAddr)
	packedBytes = append(packedBytes, addr.Bytes()...)

	nBytes, err := hex.DecodeString(nHash)
	if err != nil {
		return nil, err
	}
	packedBytes = append(packedBytes, nBytes...)
	hash := sha256.Sum256(packedBytes)

	return hash[:], nil
}

func SignData(amount *big.Int, userAddr string, nHash string) (string, error) {
	packedHash, err := packed(amount, userAddr, nHash)
	if err != nil {
		return "", err
	}
	privateKey, _, err := GetAccountByPriKey("67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e")
	if err != nil {
		return "", err
	}

	sig, err := crypto.Sign(packedHash, privateKey)
	if err != nil {
		return "", err
	}
	fmt.Println("sig1, ", hex.EncodeToString(sig))
	v := sig[len(sig)-1]
	if v == 0 || v == 1 {
		sig[len(sig)-1] = v + 27
		return hex.EncodeToString(sig), nil
	} else if v == 27 || v == 28 {
		return hex.EncodeToString(sig), nil
	} else {
		return "", fmt.Errorf("invalid signature 'v' value: %s", hex.EncodeToString(sig))
	}
}
