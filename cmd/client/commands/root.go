package commands

import (
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/ethereum/go-ethereum/ethclient"
	flag "github.com/jessevdk/go-flags"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/signer"
)

var (
	hubUrl string

	// neo setting
	neoUrl            = "http://seed3.ngd.network:20332"
	neoContract       = "cedfd8f78bf46d28ac07b8e40b911199bd51951f"
	neoContractLE     util.Uint160
	neoAssetAddr      = "Ac2EMY7wCV9Hn9LR1wMWbjgGCqtVofmd6W"
	neoUserAccount, _ = wallet.NewAccountFromWIF("L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR")
	neoUserAddr       = neoUserAccount.Address
	//neoUserAddr        = "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK"
	neoConfirmedHeight int

	// eth setting
	ethUrl             = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	ethContract        = "0x0bA64B339281D4F57DF8B535D61c6ceA71CCc956"
	ethOwnerAddress    = "0x0A8EFAacbeC7763855b9A39845DDbd03b03775C1"
	ethUserPrivate     = "aaa052c4f2eed8b96335af467b2ff80dd3a734c57d5ec4b0a8b19e1242ddc601"
	ethUserAddress     = "0xf6933949C4096670562a5E3a21B8c29c2aacA505"
	ethConfirmedHeight int
)

var (
	neoTrasaction  *neo.Transaction
	ethTransaction *eth.Transaction
	singerClient   *signer.SignerClient
	cfg            = &config.Config{}
	hubCmd         = &HubCmd{}
)

type HubCmd struct {
	SignerToken string `json:"signerToken"  long:"signerToken" description:"singer JWT token" default:"eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJRTENDaGFpbiBCb3QiLCJleHAiOjE2MzEwNjQ2MjgsImp0aSI6Ijc4MzdhNzM4LWJmYmYtNDdjNy1hZTQwLWZkMDZmY2VkZjViMSIsImlhdCI6MTU5OTUyODYyOCwiaXNzIjoiUUxDQ2hhaW4gQm90Iiwic3ViIjoic2lnbmVyIiwicm9sZXMiOlsidXNlciJdfQ.AfhfQZt-avkTC_VTtDYp8mILxEslpCncybCWi16VMKfDmPeb9zPqQylByZH9YtOvJeQSZLddQFnUfEr4yidr14lzAeoUqjdPetnN9nmNyglSqHhh3Wz_F7LmgLbaYwlvHEtSFDsKPocewHRkGzKvJCrUwxxtRcngqmcdlhp4IimNT1rZ" validate:"nonzero"`
	HubToken    string `json:"hubToken"  long:"hubToken" description:"hub JWT token"  default:"eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJRTENDaGFpbiBCb3QiLCJqdGkiOiI2ZTAwNGJiMS05ZmFiLTRjNGEtYjhiYy0yMDY3YTIzNjIyNzEiLCJpYXQiOjE1OTk1MzI2OTIsImlzcyI6IlFMQ0NoYWluIEJvdCIsInN1YiI6InNpZ25lciIsInJvbGVzIjpbInVzZXIiXX0.AZod7o926gI8r7rts8DmYTNEcJAnHl06YaoMUdt1liwnCwMXSOHZMxMNmrJ4z6-qLs7VT494Q3J14nOULKxNspkBAQ8ADObBRf110tmJiuvSENjjZe1hULM-OrnJkotzA4l81KOsotGNM3VAFTMSddPP_6RB2naAaJZDPS6sMIQGrMfv" validate:"nonzero"`
	TestNet     bool   `json:"testNet"  long:"testNet" description:"test net" `
}

func initParams(osArgs []string) {
	flag.ParseArgs(cfg, []string{})
	flag.ParseArgs(hubCmd, osArgs)
	cfg.SignerToken = hubCmd.SignerToken

	if hubCmd.TestNet {
		hubUrl = "https://hub-test.qlcchain.online"
	} else {
		hubUrl = "http://127.0.0.1:19745"
	}

	neoConfirmedHeight = cfg.NEOCfg.ConfirmedHeight
	ethConfirmedHeight = int(cfg.EthCfg.ConfirmedHeight)

	cfg.SignerEndPoint = "http://127.0.0.1:19747"
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
		ethTransaction = eth.NewTransaction(eClient, ethContract)
	}
	//defer ethClient.Close()

	log.Println("hub endpoint: ", hubUrl)
	log.Println("neo contract: ", neoContract)
	log.Println("neo endpoint: ", neoUrl)
	log.Println("eth contract: ", ethContract)
	log.Println("eth endpoint: ", ethUrl)
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
	addNeoCmd(shell)
	addHubCmd(shell)
	// run shell
	shell.Run()
}
