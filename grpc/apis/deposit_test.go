package apis

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
)

func TestDepostAPI_SignData(t *testing.T) {
	amount := big.NewInt(290000000)
	userAddr := "0xf6933949c4096670562a5e3a21b8c29c2aaca505"
	neoTx := "0x9e4ef4b8d72a4bd12851bba7aee6886afa5bfd83f57386705c8e3afae0683ead"
	priKey := "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	s, err := signData(amount, userAddr, priKey, neoTx)
	fmt.Println("sig2, ", s, err)
}

func signData(amount *big.Int, userAddr, priKey string, nHash string) (string, error) {
	packedHash, err := packed(amount, userAddr, nHash)
	if err != nil {
		return "", err
	}
	fmt.Println("packed hex: ", hex.EncodeToString(packedHash))
	privateKey, _, err := eth.GetAccountByPriKey(priKey)
	if err != nil {
		return "", err
	}

	sig, err := crypto.Sign(packedHash, privateKey)
	if err != nil {
		return "", err
	}
	fmt.Println("sig ", sig)
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

func TestNewDepositAPI_Time(t *testing.T) {
	start := time.Now().Unix()
	time.Sleep(5345 * time.Millisecond)
	end := time.Now().Unix()
	fmt.Println(end - start)
}
