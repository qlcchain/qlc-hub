package signer

import (
	"bytes"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

func TestAuthClient_SignNeoTx(t *testing.T) {
	t.Skip()
	cfg := &config.Config{
		Verbose:        true,
		LogLevel:       "debug",
		SignerToken:    "eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJRTENDaGFpbiBCb3QiLCJleHAiOjE2MzA4MjY2NjcsImp0aSI6IjQ5NjM1ZWRhLWI3YWEtNDIyYi1hMjBjLTZlNzZiMjY3MWU5YiIsImlhdCI6MTU5OTI5MDY2NywiaXNzIjoiUUxDQ2hhaW4gQm90Iiwic3ViIjoic2lnbmVyIiwicm9sZXMiOlsidXNlciJdfQ.AHU1DY4dLeL1T8LxAXDQg_iPPcWklrHsQ1hm843ykZdZ07udRtqCSztyUHKYW0iv66Xq9vUeR4odZB9qpXqQELVXAM_jAx_-o0Tc_HMXLNmnyLw9xE_VmmIMp0O6ZrYxI_Iw9Uab27od5QCAoU6fMHvbDTM46s0C2GtOy-4RXAJd8j6h",
		SignerEndPoint: "tpc://0.0.0.0:19747",
	}
	signer, err := NewSigner(cfg)
	if err != nil {
		t.Fatal(err)
	}
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
