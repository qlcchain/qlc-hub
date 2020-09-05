package signer

import (
	"testing"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
)

func TestAuthClient_SignNeoTx(t *testing.T) {
	//t.Skip()
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
}
