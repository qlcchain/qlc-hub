package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/qlc-hub/config"
)

func TestNewRPCService(t *testing.T) {
	dir := filepath.Join(config.TestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	rpc, err := NewRPCService(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	err = rpc.Init()
	if err != nil {
		t.Fatal(err)
	}
	if rpc.State() != 2 {
		t.Fatal("rpc init failed")
	}
	err = rpc.Start()
	if err != nil {
		t.Fatal(err)
	}
	if rpc.State() != 4 {
		t.Fatal("rpc start failed")
	}
	if r := rpc.RPC(); r == nil {
		t.Fatal()
	}
	err = rpc.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if rpc.Status() != 6 {
		t.Fatal("stop failed.")
	}

}
