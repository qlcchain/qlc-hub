package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

const (
	CfgFileName   = "hub.json"
	configVersion = 1
	cfgDir        = "GHub"
	nixCfgDir     = ".ghub"
)

func DefaultConfig(dir string) (*Config, error) {
	var cfg Config
	cfg = Config{
		Version:  configVersion,
		DataDir:  dir,
		LogLevel: "error",
		ChainUrl: "ws://127.0.0.1:19736",
		RPC: RPCConfig{
			Enable:             true,
			ListenAddress:      "tcp://0.0.0.0:19745",
			GRPCListenAddress:  "tcp://0.0.0.0:19746",
			CORSAllowedOrigins: []string{"*"},
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
