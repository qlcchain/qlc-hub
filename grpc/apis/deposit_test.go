package apis

import (
	"github.com/qlcchain/qlc-hub/config"
	"testing"
)

var cfg = &config.Config{
	Verbose:  false,
	LogLevel: "",
	NEOCfg: &config.NEOCfg{
		EndPoint:    "http://seed2.ngd.network:20332",
		Contract:    "0533290f35572cd06e3667653255ffd6ee6430fb",
		WIF:         "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg",
		WIFPassword: "",
	},
	EthereumCfg: &config.EthereumCfg{
		EndPoint: "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57",
		Contract: "0xCD60c41De542ebaF81040A1F50B6eFD4B1547d91",
		Account:  "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e",
	},
	RPCCfg: nil,
}

func TestDepositAPI(t *testing.T) {
	//rOrigin, rHash := util.Sha256Hash()
	//fmt.Println("hash: ", rOrigin, "==>", rHash)
	//
	//api, err := NewDepositAPI(context.Background(), cfg)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//_, err = api.Lock(context.Background(), &pb.DepositLockRequest{
	//	Nep5TxHash: "",
	//	Amount:     100000,
	//	RHash:      rHash,
	//	Addr:       "",
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//time.Sleep(5 * time.Second)
}
