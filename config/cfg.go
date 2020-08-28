package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/validator.v2"
)

const (
	cfgDir    = "Ghub"
	nixCfgDir = ".ghub"
)

type Config struct {
	Verbose     bool         `json:"verbose" short:"V" long:"verbose" description:"show verbose debug information"`
	LogLevel    string       `json:"logLevel" short:"l" long:"level" description:"log level" default:"error"` //info,warn,debug.
	NEOCfg      *NEOCfg      `json:"neo" validate:"nonnil"`
	EthereumCfg *EthereumCfg `json:"ethereum" validate:"nonnil"`
	RPCCfg      *RPCCfg      `json:"rpc" validate:"nonnil"`
}

type NEOCfg struct {
	EndPoint    string `json:"endpoint" short:"n" long:"neoUrl" description:"NEO RPC endpoint" validate:"nonzero"`
	Contract    string `json:"contract" long:"neoContract" description:"NEO staking contract address" default:"0533290f35572cd06e3667653255ffd6ee6430fb" validate:"nonzero"`
	WIF         string `json:"wif" long:"wif" description:"NEO account WIF" validate:"nonzero"`
	WIFPassword string `json:"password" long:"password" description:"NEO account password"`
}

type EthereumCfg struct {
	EndPoint string `json:"endpoint" short:"e" long:"ethereumUrl" description:"Ethereum RPC endpoint" validate:"nonzero"`
	Contract string `json:"contract" long:"ethereumContract" description:"ethereum staking contract address" default:"0x6d37597F0d9e917baeF2727ece52AEeb8B5294c7" validate:"nonzero"`
	Account  string `json:"account" long:"account" description:"Ethereum account private key" validate:"nonzero"`
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
