package config

import (
	"github.com/qlcchain/qlc-hub/pkg/jwt"
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
	Verbose        bool            `json:"verbose" short:"V" long:"verbose" description:"show verbose debug information"`
	LogLevel       string          `json:"logLevel" short:"l" long:"level" description:"log level" default:"info"` //info,warn,debug.
	SignerToken    string          `json:"signerToken"  long:"signerToken" description:"singer JWT token" default:"eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJRTENDaGFpbiBCb3QiLCJleHAiOjE2MzEwNjQ2MjgsImp0aSI6ImI2YWEwNWQ5LTk0YTEtNGUxOC05MGU2LTQ3ZGJkZWM5MjgxOSIsImlhdCI6MTU5OTUyODYyOCwiaXNzIjoiUUxDQ2hhaW4gQm90Iiwic3ViIjoic2lnbmVyIiwicm9sZXMiOlsidXNlciJdfQ.AVQAChz2g53YiIaKCip5GihIUUXNvnjLy8KN2B5DHVWOyJI2XLGYb2eVdKACT54KlM-o24NELUgbbeR9vk2a0MhbAaxqhuO4GReI8WF7IwhBvKZ-EgAzaCuCZLG9mAHeEWnTYYaZZPcjrXEZ5J93sy040S7EtjO4inHUBze0gMbEHC1f" validate:"nonzero"`
	SignerEndPoint string          `json:"signerEndPoint"  long:"signerEndPoint" description:"singer endpoint" default:"http://127.0.0.1:19747" validate:"nonzero"`
	NEOCfg         *NEOCfg         `json:"neo" validate:"nonnil"`
	EthereumCfg    *EthereumCfg    `json:"ethereum" validate:"nonnil"`
	RPCCfg         *RPCCfg         `json:"rpc" validate:"nonnil"`
	DateDir        string          `json:"dateDir" validate:"nonnil"`
	Key            string          `json:"key" short:"K" long:"key" description:"private key" default:"XoKWAwU2QhedEUr6vXuHLoiLXjkgWBoMK25edEE7YbuBxgjKhUKcFMW7n5dmky2XUQ5gQEGyCwnqYVUmS2kGebXW5pxThzRGWivZNbZaXQgXHNKFmLTV14K62AmkwHDZxQxN8bbDGdHCXq2fhgJWeU1sk3ZuAiv41gTRpnCnmgXzV8LEPTbhJk1VGzKCNyVge5eAxg1m5ziymezX7THhDq42sHFwGDFJPpvJosd5awFQrHoE8FYoKquRBqYJuwKd5Gj8ebsfhJmm3zUmqJ8kxfHyZNHEJowQ7Fxv9zdThxRAKdvMLiiYvJtQQnrWkDfwjmSZGpgWzYtFFQofq2RC7BgVRhHtYkNqb61QY1zzSjX21AHLSkNisu3fmmGiFao18LxZF2UVnVFfDpYXfwtpkjqJedT2ccR11HckyVD7nf51udvTypD64evZQpLaQdTPXGBpHM8v2drXLQLzbZDLpoSPgmegq6h5PuaQta56oT" validate:"nonzero"`
	KeyDuration    string          `json:"duration" long:"duration" default:"0s" validate:"nonzero"`
	JwtManager     *jwt.JWTManager `json:"-"`
}

type NEOCfg struct {
	EndPoint         string `json:"endpoint" short:"n" long:"neoUrl" description:"NEO RPC endpoint" default:"http://seed2.ngd.network:20332" validate:"nonzero"`
	Contract         string `json:"contract" long:"neoContract" description:"NEO staking contract address" default:"278df62f9ba1312f1e1f4b5d239f07beaa1b5b94" validate:"nonzero"`
	SignerAddress    string `json:"signerAddress" long:"neoSignerAddress" description:"NEO address to sign tx" default:"ANFnCg69c8VfE36hBhLZRrmofZ9CZU1vqZ" validate:"nonzero"`
	AssetsAddress    string `json:"assetsAddress" long:"neoAssetsAddress" description:"NEO address to keep assets" default:"Ac2EMY7wCV9Hn9LR1wMWbjgGCqtVofmd6W" validate:"nonzero"`
	ConfirmedHeight  int    `json:"neoConfirmedHeight" long:"neoConfirmedHeight" description:"Neo transaction Confirmed Height" default:"1" validate:"nonzero"`
	DepositInterval  int64  `json:"neoDepositInterval" long:"neoDepositInterval" description:"Lock timeout interval height of deposit" default:"40" validate:"nonzero"`
	WithdrawInterval int64  `json:"neoWithdrawInterval" long:"neoWithdrawInterval" description:"Lock timeout interval height of withdraw" default:"20" validate:"nonzero"`
}

type EthereumCfg struct {
	EndPoint         string `json:"endpoint" short:"e" long:"ethereumUrl" description:"Ethereum RPC endpoint" default:"wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57" validate:"nonzero"`
	Contract         string `json:"contract" long:"ethereumContract" description:"ethereum staking contract address" default:"0x9a36F711133188EDb3952b3A6ee29c6a3d2e3836" validate:"nonzero"`
	SignerAddress    string `json:"signerAddress" long:"ethSignerAddress" description:"Ethereum address to sign tx" default:"0x0A8EFAacbeC7763855b9A39845DDbd03b03775C1" validate:"nonzero"`
	ConfirmedHeight  int    `json:"ethConfirmedHeight" long:"ethConfirmedHeight" description:"Eth transaction Confirmed Height" default:"1" validate:"nonzero"`
	DepositInterval  int64  `json:"ethDepositHeight" long:"ethDepositHeight" description:"Lock timeout Height of deposit" default:"20" validate:"nonzero"`
	WithdrawInterval int64  `json:"ethWithdrawHeight" long:"ethWithdrawHeight" description:"Lock timeout Height of withdraw" default:"40" validate:"nonzero"`
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
	if err := validator.Validate(c); err != nil {
		return err
	}
	d, err := time.ParseDuration(c.KeyDuration)
	if err != nil {
		return err
	}

	jwt, err := jwt.NewJWTManager(c.Key, d)
	if err != nil {
		return err
	}
	c.JwtManager = jwt
	return nil
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
