package config

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func TestSignerConfig_Verify(t *testing.T) {
	jwtKey := jwt.NewBase58Key()
	var neoKeys, ethKeys []string

	for i := 0; i < 2; i++ {
		key, _ := crypto.GenerateKey()
		privateKeyBytes := crypto.FromECDSA(key)
		s := hex.EncodeToString(privateKeyBytes)
		t.Logf("0:%s", s)
		ethKeys = append(ethKeys, s)
	}

	for i := 0; i < 2; i++ {
		account, _ := wallet.NewAccount()
		s := account.PrivateKey().String()
		t.Logf("1:%s", s)
		neoKeys = append(neoKeys, s)
	}

	cfg := &SignerConfig{
		Verbose:           false,
		Key:               jwtKey,
		KeyDuration:       "8760h0m0s",
		LogLevel:          "debug",
		NeoAccounts:       neoKeys,
		EthAccounts:       ethKeys,
		GRPCListenAddress: "tcp://0.0.0.0:19747",
		JwtManager:        nil,
		Keys:              nil,
	}
	if err := cfg.Verify(); err != nil {
		t.Fatal(err)
	}
	r1 := cfg.AddressList(proto.SignType_NEO)
	t.Log("NEO: ", util.ToIndentString(r1))

	r2 := cfg.AddressList(proto.SignType_ETH)
	t.Log("NEO: ", util.ToIndentString(r2))
}
