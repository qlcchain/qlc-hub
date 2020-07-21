package services

import (
	"path/filepath"
	"testing"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/services/context"
)

func TestRegisterServices(t *testing.T) {
	cfgFile2 := filepath.Join(config.TestDataDir(), "config1", config.CfgFileName)
	cm := config.NewCfgManagerWithName(filepath.Dir(cfgFile2), filepath.Base(cfgFile2))
	cc := context.NewServiceContext(cm.ConfigFile)
	if err := RegisterServices(cc); err != nil {
		t.Fatal(err)
	}

	if services, err := cc.AllServices(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(len(services))
	}
}
