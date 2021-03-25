package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/jwt"
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
	SignerToken    string          `json:"signerToken"  long:"signerToken" description:"singer JWT token" validate:"nonzero"`
	SignerEndPoint string          `json:"signerEndPoint"  long:"signerEndPoint" description:"singer endpoint" validate:"nonzero"`
	NEOCfg         *NEOCfg         `json:"neo" validate:"nonnil"`
	EthCfg         *EthCfg         `json:"ethereum" validate:"nonnil"`
	QlcCfg         *QlcCfg         `json:"qlc" validate:"nonnil"`
	BscCfg         *BSCCfg         `json:"bsc" validate:"nonnil"`
	RPCCfg         *RPCCfg         `json:"rpc" validate:"nonnil"`
	DateDir        string          `json:"dateDir" validate:"nonnil"`
	CanRefund      int             `json:"canRefund" long:"canRefund" description:"deposit can refund"  default:"0"`
	Key            string          `json:"key" short:"K" long:"key" description:"private key" validate:"nonzero"`
	KeyDuration    string          `json:"duration" long:"duration" default:"0s" validate:"nonzero"`
	JwtManager     *jwt.JWTManager `json:"-"`
}

type NEOCfg struct {
	EndPoints       []string `json:"neoUrls"  long:"neoUrls" description:"NEO RPC endpoint" validate:"min=1"`
	Contract        string   `json:"contract" long:"neoContract" description:"NEO staking contract address" validate:"nonzero"`
	OwnerAddress    string   `json:"neoOwnerAddress" long:"neoOwnerAddress" description:"NEO address to sign tx" validate:"nonzero"`
	ConfirmedHeight int      `json:"neoConfirmedHeight" long:"neoConfirmedHeight" description:"Neo transaction Confirmed Height" default:"1" validate:""`
}

type EthCfg struct {
	EndPoints        []string `json:"ethUrls" long:"ethUrls" description:"Ethereum RPC endpoint" validate:"nonzero"`
	Nep5Contract     string   `json:"contract" long:"nep5Contract" description:"QLC staking contract address"  validate:"nonzero"`
	QGasSwapContract string   `json:"contract" long:"qGasSwapContract" description:"QGas Swap contract address"  validate:"nonzero"`
	OwnerAddress     string   `json:"ethOwnerAddress" long:"ethOwnerAddress" description:"Ethereum owner address"  validate:"nonzero"`
	ConfirmedHeight  int64    `json:"ethConfirmedHeight" long:"ethConfirmedHeight" description:"Eth transaction Confirmed Height" default:"3" validate:""`
}

type BSCCfg struct {
	EndPoints       []string `json:"bscUrls" long:"bscUrls" description:"BSC RPC endpoint" validate:"nonzero"`
	Contract        string   `json:"contract" long:"bscContract" description:"BSC staking contract address"  validate:"nonzero"`
	OwnerAddress    string   `json:"bscOwnerAddress" long:"bscOwnerAddress" description:"BSC owner address"  validate:"nonzero"`
	ConfirmedHeight int64    `json:"bscConfirmedHeight" long:"bscConfirmedHeight" description:"BSC transaction Confirmed Height" default:"3" validate:""`
}

type QlcCfg struct {
	EndPoint     string `json:"qlcUrl" long:"qlcUrl" description:"QLC RPC endpoint" validate:"nonzero"`
	OwnerAddress string `json:"qlcOwnerAddress" long:"qlcOwnerAddress" description:"qlc owner address"  validate:"nonzero"`
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

func (c *Config) Database() string {
	dir := filepath.Join(DefaultDataDir(), "db")
	_ = util.CreateDirIfNotExist(dir)
	return filepath.Join(dir, "swap.db")
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
