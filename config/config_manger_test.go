package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/qlc-hub/common/util"
)

var configDir = filepath.Join(TestDataDir(), "config")

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	return func(t *testing.T) {
		t.Log("teardown test case")
		err := os.RemoveAll(configDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestConfigManager_Load(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	manager := NewCfgManager(configDir)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
}

func TestConfigManager_parseVersion(t *testing.T) {
	manager := NewCfgManager(configDir)
	cfg, err := DefaultConfig(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if version, err := manager.parseVersion(bytes); err != nil {
		t.Fatal(err)
	} else {
		t.Log(version)
	}
}

func TestNewCfgManagerWithFile(t *testing.T) {
	cfgFile := filepath.Join(configDir, "test.json")
	defer func() {
		_ = os.Remove(cfgFile)
	}()
	cm := NewCfgManagerWithFile(cfgFile)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	} else {
		_ = cm.Save()
	}
}

func TestDefaultConfig(t *testing.T) {
	manager := NewCfgManager(configDir)
	cfg, err := DefaultConfig(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
}
