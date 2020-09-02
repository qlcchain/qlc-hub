package commands

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"go.uber.org/zap"
)

var (
	// hub
	//hubUrl = "https://hub-test.qlcchain.online"
	hubUrl = "http://127.0.0.1:19745"

	// neo setting
	neoUrl        = "http://seed2.ngd.network:20332"
	neoContract   string
	neoContractLE util.Uint160

	neoUserWif           = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	neoUserAccount, _    = wallet.NewAccountFromWIF(neoUserWif)
	neoWrapperWif        = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	neoWrapperAccount, _ = wallet.NewAccountFromWIF(neoWrapperWif)
	userEthAddress       = "2e1ac6242bb084029a9eb29dfb083757d27fced4"

	// eth setting
	ethUrl                  = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	ethContract             string
	ethWrapperPrikey        = "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	_, ethWrapperAccount, _ = eth.GetAccountByPriKey(ethWrapperPrikey)
	ethUserPrikey           = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
	userEthPrikey           = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"

	ethIntervalHeight = 20
	neoIntervalHeight = 20
)

var (
	neoTrasaction *neo.Transaction
	ethClient     *ethclient.Client
	lockAmount    = 130000000
	logger        *zap.SugaredLogger
)

func init() {
	var err error
	logger = log.NewLogger("api/debug")
	ethContract, neoContract = getContractAddress()
	neoContractLE, err = util.Uint160DecodeStringLE(neoContract)
	if err != nil {
		logger.Fatal(err)
	}

	neoTrasaction, err = neo.NewTransaction(neoUrl, neoContract)
	if err != nil {
		logger.Fatal(err)
	}

	ethClient, err = ethclient.Dial(ethUrl)
	if err != nil {
		logger.Fatal(err)
	}
	//defer ethClient.Close()

	logger.Info("neo contract: ", neoContract)
	logger.Info("eth contract: ", ethContract)
}
