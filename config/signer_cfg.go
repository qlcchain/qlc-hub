package config

import (
	"crypto/ecdsa"
	"errors"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/validator.v2"

	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/log"
)

const (
	signerDir       = "Signer"
	nixSignerCfgDir = ".signer"
)

type SignerConfig struct {
	Verbose           bool                                      `json:"verbose" short:"V" long:"verbose" description:"show verbose debug information"`
	Key               string                                    `json:"key" short:"K" long:"key" description:"private key" validate:"nonzero"`
	KeyDuration       string                                    `json:"duration" long:"duration" default:"8760h0m0s" validate:"nonzero"`
	LogLevel          string                                    `json:"logLevel" short:"l" long:"level" description:"log level" default:"warn"` //info,warn,debug.
	ChainAccounts     map[int][]string                          `json:"keys" long:"accounts" description:"NEO/ETH private key" validate:"min=1"`
	GRPCListenAddress string                                    `json:"gRPCListenAddress" long:"grpcAddress" description:"GRPC server listen address" default:"tcp://0.0.0.0:19747"`
	JwtManager        *jwt.JWTManager                           `json:"-"`
	Keys              map[proto.SignType]map[string]interface{} `json:"-"`
}

func (c *SignerConfig) Verify() error {
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

	for k, priKeys := range c.ChainAccounts {
		t := proto.SignType(k)
		for _, v := range priKeys {
			switch t {
			case proto.SignType_ETH:
				if privateKey, err := crypto.HexToECDSA(v); err == nil {
					publicKey := privateKey.Public()
					if publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey); ok {
						address := crypto.PubkeyToAddress(*publicKeyECDSA).String()
						c.saveKey(t, address, privateKey)
					} else {
						log.Root.Error("invalid public key")
					}
				} else {
					log.Root.Errorf("can not decode private key(%s), err: %s", v, err)
				}
			case proto.SignType_NEO:
				if priv, err := keys.NewPrivateKeyFromHex(v); err == nil {
					c.saveKey(t, priv.Address(), priv)
				} else {
					log.Root.Errorf("can not decode wif key(%s),err: %s", v, err)
				}
			}
		}
	}
	if len(c.Keys) == 0 {
		return errors.New("can not find any invalid keys")
	}

	return nil
}

func (c *SignerConfig) LogDir() string {
	return filepath.Join(defaultDataDir(), "log", time.Now().Format("2006-01-02T15-04"))
}

func (c *SignerConfig) saveKey(t proto.SignType, address string, privateKey interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[proto.SignType]map[string]interface{}, 0)
	}
	if _, ok := c.Keys[t]; !ok {
		c.Keys[t] = make(map[string]interface{}, 0)
	}
	c.Keys[t][address] = privateKey
}

func (c *SignerConfig) AddressList(t proto.SignType) []string {
	var result []string
	if vals, ok := c.Keys[t]; ok {
		for k, _ := range vals {
			result = append(result, k)
		}
	}
	return result
}

func defaultDataDir() string {
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", signerDir)
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", signerDir)
		} else {
			return filepath.Join(home, nixSignerCfgDir)
		}
	}
	return ""
}
