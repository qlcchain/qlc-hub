package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

const (
	cfgDir    = "Ghub"
	nixCfgDir = ".ghub"
)

type Config struct {
	Verbose           bool            `json:"verbose" short:"V" long:"verbose" description:"show verbose debug information"`
	LogLevel          string          `json:"logLevel" short:"l" long:"level" description:"log level" default:"info"` //info,warn,debug.
	SignerToken       string          `json:"signerToken"  long:"signerToken" description:"singer JWT token" validate:"nonzero"`
	SignerEndPoint    string          `json:"signerEndPoint"  long:"signerEndPoint" description:"singer endpoint" validate:"nonzero"`
	NEOCfg            *NEOCfg         `json:"neo" validate:"nonnil"`
	EthereumCfg       *EthereumCfg    `json:"ethereum" validate:"nonnil"`
	RPCCfg            *RPCCfg         `json:"rpc" validate:"nonnil"`
	DateDir           string          `json:"dateDir" validate:"nonnil"`
	MinDepositAmount  int64           `json:"minDepositAmount" long:"minDepositAmount" description:"minimal amount to deposit" default:"100000000" validate:"nonzero"`
	MinWithdrawAmount int64           `json:"minWithdrawAmount" long:"minWithdrawAmount" description:"minimal amount to withdraw" default:"100000000" validate:"nonzero"`
	WithdrawFrequency int             `json:"withdrawFrequency" long:"withdrawFrequency" description:"time interval to every withdraw (minute)" default:"10" validate:"nonzero"`
	StateInterval     int             `json:"stateInterval" long:"stateInterval" description:"time interval to check locker state" default:"2" validate:"nonzero"`
	Key               string          `json:"key" short:"K" long:"key" description:"private key" validate:"nonzero"`
	KeyDuration       string          `json:"duration" long:"duration" default:"0s" validate:"nonzero"`
	JwtManager        *jwt.JWTManager `json:"-"`
}

type NEOCfg struct {
	EndPoint         string `json:"endpoint" short:"n" long:"neoUrl" description:"NEO RPC endpoint" validate:"nonzero"`
	Contract         string `json:"contract" long:"neoContract" description:"NEO staking contract address" validate:"nonzero"`
	AssetId          string `json:"assetId" long:"neoAssetId" description:"qlc token asset id"  validate:"nonzero"`
	SignerAddress    string `json:"signerAddress" long:"neoSignerAddress" description:"NEO address to sign tx" validate:"nonzero"`
	AssetsAddress    string `json:"assetsAddress" long:"neoAssetsAddress" description:"NEO address to keep assets" validate:"nonzero"`
	ConfirmedHeight  int    `json:"neoConfirmedHeight" long:"neoConfirmedHeight" description:"Neo transaction Confirmed Height" default:"0" validate:""`
	DepositInterval  int64  `json:"neoDepositInterval" long:"neoDepositInterval" description:"Lock timeout interval height of deposit" default:"40" validate:"nonzero"`
	WithdrawInterval int64  `json:"neoWithdrawInterval" long:"neoWithdrawInterval" description:"Lock timeout interval height of withdraw" default:"20" validate:"nonzero"`
}

type EthereumCfg struct {
	EndPoint         string `json:"endpoint" short:"e" long:"ethereumUrl" description:"Ethereum RPC endpoint" validate:"nonzero"`
	Contract         string `json:"contract" long:"ethereumContract" description:"ethereum staking contract address"  validate:"nonzero"`
	OwnerAddress     string `json:"ethOwnerAddress" long:"ethOwnerAddress" description:"Ethereum owner address" validate:"nonzero"`
	ConfirmedHeight  int    `json:"ethConfirmedHeight" long:"ethConfirmedHeight" description:"Eth transaction Confirmed Height" default:"0" validate:""`
	DepositInterval  int64  `json:"ethDepositHeight" long:"ethDepositHeight" description:"Lock timeout Height of deposit" default:"20" validate:"nonzero"`
	WithdrawInterval int64  `json:"ethWithdrawHeight" long:"ethWithdrawHeight" description:"Lock timeout Height of withdraw" default:"40" validate:"nonzero"`
	GasEndPoint      string `json:"gasEndPoint" long:"gasEndPoint" description:"endpoint to get gas price" validate:"nonzero"`
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

	jm, err := jwt.NewJWTManager(c.Key, d)
	if err != nil {
		return err
	}
	c.JwtManager = jm
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
