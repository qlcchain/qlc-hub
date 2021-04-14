package commands

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	flag "github.com/jessevdk/go-flags"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/signer"
)

var (
	hubUrl string

	// neo setting
	neoUrl            = []string{"http://seed5.ngd.network:20332"}
	neoContract       = "bfcbb52d61bc6d3ef2c8cf43f595f4bf5cac66c5"
	neoContractLE     util.Uint160
	neoOwnerAddress   = "ANFnCg69c8VfE36hBhLZRrmofZ9CZU1vqZ"
	neoUserWif        = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	neoUserAccount, _ = wallet.NewAccountFromWIF(neoUserWif)
	neoUserAddr       = neoUserAccount.Address
	//neoUserAddr        = "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK"
	neoConfirmedHeight int

	// eth setting (nep5 -> eth)
	ethUrl             = []string{"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"}
	ethContract        = "0xE2484A4178Ce7FfD5cd000030b2a5de08c0Caf8D"
	ethOwnerAddress    = "0x0A8EFAacbeC7763855b9A39845DDbd03b03775C1"
	ethUserPrivate     = "aaa052c4f2eed8b96335af467b2ff80dd3a734c57d5ec4b0a8b19e1242ddc601"
	ethUserAddress     = "0xf6933949C4096670562a5E3a21B8c29c2aacA505"
	ethConfirmedHeight int

	// bsc setting (nep5 -> bsc)
	bscUrl          = []string{"https://bsc-rpc-test.qlcchain.online"}
	bscContractNep5 = "0xF284c1C1D03BCCC8b32e9736919C0D7CA8b06aeD"
	bscUserPrivate  = "d6aedb156c57320b7246c4463c9ee9c9d54df23513ece5eda0f2c9d3fdfc5822"
	bscUserAddress  = "0x318c6E6613D34a57972f2679d5039E807f048C6E"

	// eth setting (qgas -> eth)
	ethContractQGas = "0x334856843E77B9f1a857814c96289236F13418D9"

	// bsc setting (qgas -> eth)
	bscContractQGas = "0xfEF38Cebfa3c73a194b1296f5c6dbaafc81f77Da"

	// qlc setting
	qlcUserPrivate = "8be0696a2d51dec8e2859dcb8ce2fd7ce7412eb9d6fa8a2089be8e8f1eeb4f0e458779381a8d21312b071729344a0cb49dc1da385993e19d58b5578da44c0df0"
	priv, _        = hex.DecodeString(qlcUserPrivate)
	qlcUserAccount = qlctypes.NewAccount(priv)
	qlcUserAddress = "qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z"
)

var (
	neoTrasaction      *neo.Transaction
	ethTransactionNep5 *eth.Transaction
	bscTransactionNep5 *eth.Transaction
	ethTransactionQLC  *eth.Transaction
	bscTransactionQLC  *eth.Transaction
	singerClient       *signer.SignerClient
	cfg                = &config.Config{}
	hubCmd             = &HubCmd{}
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
	ethConfirmedHeight = int(cfg.EthCfg.EthConfirmedHeight)
	var err error
	if neoContractLE, err = util.Uint160DecodeStringLE(neoContract); err != nil {
		log.Fatal(err)
	}

	if neoTrasaction, err = neo.NewTransaction(neoUrl, neoContract, nil); err != nil {
		log.Fatal(err)
	}
	if c := neoTrasaction.Client(); c == nil {
		log.Fatal("invalid neo endpoints")
	}

	ethTransactionNep5, _ = eth.NewTransaction(ethUrl, ethContract)
	bscTransactionNep5, _ = eth.NewTransaction(bscUrl, bscContractNep5)
	ethTransactionQLC, _ = eth.NewTransaction(ethUrl, ethContractQGas)
	bscTransactionQLC, _ = eth.NewTransaction(bscUrl, bscContractQGas)
	//defer ethClient.Close()

	log.Println("hub endpoint: ", hubUrl)
	log.Println("neo contract: ", neoContract)
	log.Println("neo endpoint: ", neoUrl)
	log.Println("neo user address: ", neoUserAddr)
	log.Println("eth contract: ", ethContract)
	log.Println("eth endpoint: ", ethUrl)
	log.Println("eth user address: ", ethUserAddress)
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
	addQLCCmd(shell)
	// run shell
	shell.Run()
}
