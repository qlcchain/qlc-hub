package client

import (
	"testing"

	"github.com/qlcchain/qlc-hub/grpc/proto"
)

func TestAuthClient_SignNeoTx(t *testing.T) {
	t.Log(proto.SignType_ETH.Number())
}
