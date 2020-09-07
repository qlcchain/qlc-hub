package commands

import (
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/ethereum/go-ethereum/ethclient"
	flag "github.com/jessevdk/go-flags"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/signer"
)

var (
	// hub
	//hubUrl = "https://hub-test.qlcchain.online"
	hubUrl = "http://127.0.0.1:19745"

	// neo setting
	neoUrl        string
	neoContract   string
	neoContractLE util.Uint160

	//neoUserWif          = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	//neoUserAccount, _   = wallet.NewAccountFromWIF(neoUserWif)
	neoUserAddr             = "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK"
	neoWrapperAssetAddr     string
	neoWrapperSignerAddress string
	userEthAddress          = "2e1ac6242bb084029a9eb29dfb083757d27fced4"

	// eth setting
	ethUrl                 string
	ethContract            string
	ethWrapperOwnerAddress string // also eth wrapper signer address
	//ethUserPrikey           = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
	ethUserAddress = "0x6A786bf6E1c68E981D04139137f81dDA2d0acBF1"

	ethIntervalHeight = 20
	neoIntervalHeight = 20
)

var (
	neoTrasaction  *neo.Transaction
	ethTransaction *eth.Transaction
	lockAmount     = 130000000
	singerClient   *signer.SignerClient
	cfg            = &config.Config{}
)

func initParams(osArgs []string) {
	flag.ParseArgs(cfg, osArgs)

	neoUrl = cfg.NEOCfg.EndPoint
	neoContract = cfg.NEOCfg.Contract
	neoWrapperSignerAddress = cfg.NEOCfg.SignerAddress
	neoWrapperAssetAddr = cfg.NEOCfg.AssetsAddress

	ethUrl = cfg.EthereumCfg.EndPoint
	ethContract = cfg.EthereumCfg.Contract
	ethWrapperOwnerAddress = cfg.EthereumCfg.SignerAddress

	var err error
	if singerClient, err = signer.NewSigner(cfg); err != nil {
		log.Fatal(err)
	}

	if neoContractLE, err = util.Uint160DecodeStringLE(neoContract); err != nil {
		log.Fatal(err)
	}

	if neoTrasaction, err = neo.NewTransaction(neoUrl, neoContract, singerClient); err != nil {
		log.Fatal(err)
	}
	if err := neoTrasaction.Client().Ping(); err != nil {
		log.Fatal(err)
	}

	if eClient, err := ethclient.Dial(ethUrl); err != nil {
		log.Fatal(err)
	} else {
		ethTransaction = eth.NewTransaction(eClient, singerClient, ethContract)
	}
	//defer ethClient.Close()

	log.Println("neo contract: ", neoContract)
	log.Println("eth contract: ", ethContract)
}

func Execute(osArgs []string) {
	initParams(osArgs)
	shell := ishell.NewWithConfig(&readline.Config{
		Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
		HistoryFile: "/tmp/readline.tmp",
		//AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		//FuncFilterInputRune: filterInput,
	})
	shell.Println("QLC Hub Client")
	//set common variable
	addEthCmd(shell)
	addNeoCmd(shell)
	addHubCmd(shell)
	// run shell
	shell.Run()
}
