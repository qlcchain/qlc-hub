package config

import (
	"encoding/base64"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	CfgFileName        = "hub.json"
	configVersion      = 1
	cfgDir             = "GHub"
	nixCfgDir          = ".ghub"
	bootNodeHttpServer = "0.0.0.0:9997"
)

var bootNodes = []string{"https://seed1-hub.qlcchain.online/bootNode", "https://seed2-hub.qlcchain.online/bootNode"}

// identityConfig initializes a new identity.
func identityConfig() (string, string, error) {
	sk, pk, err := ic.GenerateKeyPair(ic.RSA, 2048)
	if err != nil {
		return "", "", err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	if err != nil {
		return "", "", err
	}
	privKey := base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", "", err
	}
	peerID := id.Pretty()
	return privKey, peerID, nil
}

func DefaultConfig(dir string) (*Config, error) {
	pk, id, err := identityConfig()
	if err != nil {
		return nil, err
	}
	var cfg Config
	cfg = Config{
		Version:  configVersion,
		DataDir:  dir,
		LogLevel: "error",
		ChainUrl: "ws://127.0.0.1:19736",
		RPC: RPCConfig{
			Enable:             true,
			ListenAddress:      "tcp://0.0.0.0:9998",
			GRPCListenAddress:  "tcp://0.0.0.0:9999",
			CORSAllowedOrigins: []string{"*"},
		},
		P2P: &P2PConfig{
			BootNodes:          bootNodes,
			IsBootNode:         false,
			BootNodeHttpServer: bootNodeHttpServer,
			Listen:             "/ip4/0.0.0.0/tcp/9996",
			ListeningIp:        "127.0.0.1",
			SyncInterval:       120,
			Discovery: &DiscoveryConfig{
				DiscoveryInterval: 60,
				Limit:             2000,
				MDNSEnabled:       true,
				MDNSInterval:      30,
			},
			ID: &IdentityConfig{id, pk},
		},
	}
	return &cfg, nil
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

func (c *Config) LogDir() string {
	return filepath.Join(c.DataDir, "log", time.Now().Format("2006-01-02T15-04"))
}

func TestDataDir() string {
	return filepath.Join(DefaultDataDir(), "test")
}

// DecodePrivateKey is a helper to decode the users PrivateKey
func (c *Config) DecodePrivateKey() (ic.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(c.P2P.ID.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO:(security)
	return ic.UnmarshalPrivateKey(pkb)
}
