package apis

import (
	"context"
	"testing"
)

var url = "https://ethgasstation.info/api/ethgasAPI.json?api-key=dcc85335d8be462feedfc78fa4f69536a953b37b7942aca02b044c1e0816"

func TestGetBestGas(t *testing.T) {
	GetBestGas(context.Background(), url)
}
