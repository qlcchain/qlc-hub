package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/util"
	"gopkg.in/validator.v2"
)

const (
	cfgDir    = "Ghub"
	nixCfgDir = ".ghub"
)

type Config struct {
	Verbose     bool         `json:"verbose" short:"V" long:"verbose" description:"show verbose debug information"`
	LogLevel    string       `json:"logLevel" short:"l" long:"level" description:"log level" default:"info"` //info,warn,debug.
	NEOCfg      *NEOCfg      `json:"neo" validate:"nonnil"`
	EthereumCfg *EthereumCfg `json:"ethereum" validate:"nonnil"`
	RPCCfg      *RPCCfg      `json:"rpc" validate:"nonnil"`
	DateDir     string       `json:"dateDir" validate:"nonnil"`
}

type NEOCfg struct {
	EndPoint        string `json:"endpoint" short:"n" long:"neoUrl" description:"NEO RPC endpoint" default:"http://seed2.ngd.network:20332" validate:"nonzero"`
	Contract        string `json:"contract" long:"neoContract" description:"NEO staking contract address" default:"e0abb5fde5a0b870c13f3e60258856e38a939187" validate:"nonzero"`
	WIF             string `json:"wif" long:"wif" description:"NEO account WIF" default:"L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg" validate:"nonzero"`
	WIFPassword     string `json:"password" long:"password" description:"NEO account password"`
	ConfirmedHeight int    `json:"neoConfirmedHeight" long:"neoConfirmedHeight" description:"Neo transaction Confirmed Height" default:"1" validate:"nonzero"`
	DepositHeight   int64  `json:"depositNeoTimeoutHeight" long:"depositNeoTimeoutHeight" description:"Lock timeout Height of deposit" default:"40" validate:"nonzero"`
	WithdrawHeight  int64  `json:"withdrawNeoTimeoutHeight" long:"withdrawNeoTimeoutHeight" description:"Lock timeout Height of withdraw" default:"20" validate:"nonzero"`
}

type EthereumCfg struct {
	EndPoint        string `json:"endpoint" short:"e" long:"ethereumUrl" description:"Ethereum RPC endpoint" default:"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57" validate:"nonzero"`
	Contract        string `json:"contract" long:"ethereumContract" description:"ethereum staking contract address" default:"0x9a36F711133188EDb3952b3A6ee29c6a3d2e3836" validate:"nonzero"`
	Account         string `json:"account" long:"account" description:"Ethereum account private key" default:"67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e" validate:"nonzero"`
	ConfirmedHeight int    `json:"ethConfirmedHeight" long:"ethConfirmedHeight" description:"Eth transaction Confirmed Height" default:"1" validate:"nonzero"`
	DepositHeight   int64  `json:"depositEthTimeoutHeight" long:"depositEthTimeoutHeight" description:"Lock timeout Height of deposit" default:"20" validate:"nonzero"`
	WithdrawHeight  int64  `json:"withdrawEthTimeoutHeight" long:"withdrawEthTimeoutHeight" description:"Lock timeout Height of withdraw" default:"40" validate:"nonzero"`
}

type RPCCfg struct {
	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"listenAddress" long:"listenAddress" description:"RPC server listen address" default:"tcp://0.0.0.0:19745"`
	// TCP or UNIX socket address for the gRPC server to listen on
	GRPCListenAddress  string   `json:"gRPCListenAddress" long:"grpcAddress" description:"GRPC server listen address" default:"tcp://0.0.0.0:19746"`
	CORSAllowedOrigins []string `json:"allowedOrigins" long:"allowedOrigins" description:"AllowedOrigins of CORS" default:"*"`
}

func (c *Config) LogDir() string {
	return filepath.Join(DefaultDataDir(), "log", time.Now().Format("2006-01-02T15-04"))
}

func (c *Config) DataDir() string {
	dir := filepath.Join(DefaultDataDir(), "data")
	_ = util.CreateDirIfNotExist(dir)

	return dir
}

func (c *Config) Verify() error {
	return validator.Validate(c)
}

// DefaultDataDir is the default data directory to use for the databases and other persistence requirements.
func DefaultDataDir() string {
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", cfgDir)
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", cfgDir)
		} else {
			return filepath.Join(home, nixCfgDir)
		}
	}
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
