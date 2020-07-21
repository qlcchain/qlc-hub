package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/qlc-hub/config"
)

func TestLogService(t *testing.T) {
	cfgFile2 := filepath.Join(config.TestDataDir(), "log", config.CfgFileName)
	cm := config.NewCfgManagerWithName(filepath.Dir(cfgFile2), filepath.Base(cfgFile2))
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := cm.ParseDataDir()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ps := NewLogService(cm.ConfigFile)

	err = ps.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ps.State() != 2 {
		t.Fatal("log service init failed")
	}
	err = ps.Start()
	if err != nil {
		t.Fatal(err)
	}
	_ = ps.Stop()

	if ps.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
