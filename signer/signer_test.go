package signer

import (
	"bytes"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"

	"github.com/qlcchain/qlc-hub/grpc"
	"github.com/qlcchain/qlc-hub/pkg/jwt"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func TestAuthClient_SignNeoTx(t *testing.T) {
	t.Skip()
	jwtKey := jwt.NewBase58Key()
	signerAddr := "tpc://0.0.0.0:19747"

	signerCfg := &config.SignerConfig{
		Verbose:           false,
		Key:               jwtKey,
		KeyDuration:       "0s",
		LogLevel:          "debug",
		NeoAccounts:       []string{"1586faf62462d7b87c12c4c98ed28042b2bdfd715d3040baccdc267dfd60859a", "51ddd7e66c6b6ce0baf07e3aa52e15d20ef8238bf73b05a4b1b4aa1e4f13bbb9"},
		EthAccounts:       []string{"5d5f13593918431c70354607060d67e931a8bdc0575b4328e8ebb367b0d86d1d", "d38fbafe777c3d49e35e708961fb43407bca9d804ee2d86b22f8335914745998"},
		GRPCListenAddress: signerAddr,
		JwtManager:        nil,
		Keys:              nil,
	}

	if err := signerCfg.Verify(); err != nil {
		t.Fatal(err)
	}

	signerServer, err := grpc.NewSignerServer(signerCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		signerServer.Stop()
	}()

	token, err := signerCfg.JwtManager.Generate(jwt.User)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Verbose:        true,
		LogLevel:       "debug",
		SignerToken:    token,
		SignerEndPoint: signerAddr,
	}
	signer, err := NewSigner(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		signer.Stop()
	}()

	if r1, err := signer.AddressList(proto.SignType_NEO); err == nil {
		t.Log(r1.Address)
	} else {
		t.Fatal(err)
	}
	if r1, err := signer.AddressList(proto.SignType_ETH); err == nil {
		t.Log(r1.Address)
	} else {
		t.Fatal(err)
	}

	rawData := []byte(util.String(120))
	neo := "1586faf62462d7b87c12c4c98ed28042b2bdfd715d3040baccdc267dfd60859a"
	neoPriv, _ := keys.NewPrivateKeyFromHex(neo)
	neoAddress := neoPriv.Address()
	sign := neoPriv.Sign(rawData)
	if neoResp, err := signer.Sign(proto.SignType_NEO, neoAddress, rawData); err == nil {
		if !bytes.Equal(sign, neoResp.Sign) {
			t.Fatalf("got: %v, exp: %v", neoResp.Sign, sign)
		}
	} else {
		t.Fatal(err)
	}

	eth := "5d5f13593918431c70354607060d67e931a8bdc0575b4328e8ebb367b0d86d1d"
	ethPriv, _ := crypto.HexToECDSA(eth)
	publicKey := ethPriv.Public().(*ecdsa.PublicKey)
	ethAddress := crypto.PubkeyToAddress(*publicKey).String()
	h := crypto.Keccak256(rawData)
	ethSign, _ := crypto.Sign(h, ethPriv)

	if ethResp, err := signer.Sign(proto.SignType_ETH, ethAddress, h); err == nil {
		if !bytes.Equal(ethSign, ethResp.Sign) {
			t.Fatalf("got: %v, exp: %v", ethResp.Sign, sign)
		}
	} else {
		t.Fatal(err)
	}
}
