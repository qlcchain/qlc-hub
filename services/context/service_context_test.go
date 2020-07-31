/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package context

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/qlcchain/qlc-go-sdk/pkg/types"

	"github.com/google/uuid"

	"github.com/qlcchain/qlc-hub/common"
	"github.com/qlcchain/qlc-hub/config"
)

type testService struct {
	Id int
}

func (*testService) Init() error {
	panic("implement me")
}

func (*testService) Start() error {
	return nil
}

func (*testService) Stop() error {
	return nil
}

func (*testService) Status() int32 {
	panic("implement me")
}

type waitService struct {
	common.ServiceLifecycle
}

func (w *waitService) Init() error {
	if !w.PreInit() {
		return errors.New("waitService pre init fail")
	}
	defer w.PostInit()
	return nil
}

func (w *waitService) Start() error {
	if !w.PreStart() {
		return errors.New("waitService pre start fail")
	}
	defer w.PostStart()

	time.Sleep(time.Duration(3) * time.Second)
	return nil
}

func (w *waitService) Stop() error {
	if !w.PreStop() {
		return errors.New("waitService pre stop fail")
	}
	defer func() {
		w.PostStop()
		w.Reset()
	}()

	return nil
}

func (w *waitService) Status() int32 {
	return w.State()
}

func Test_serviceContainer(t *testing.T) {
	sc := newServiceContainer()
	serv1 := &testService{Id: 1}
	t.Logf("serv1 %p", serv1)
	err := sc.Register(LedgerService, serv1)
	if err != nil {
		t.Fatal(err)
	}
	err = sc.Register(LedgerService, serv1)
	if err == nil {
		t.Fatal(err)
	}

	if service, err := sc.Get(LedgerService); service == nil || err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%p, %p", service, serv1)
	}

	if b := sc.HasService(LedgerService); !b {
		t.Fatal("can not find ledger service")
	}

	serv2 := &testService{Id: 2}
	t.Logf("serv2 %p", serv2)
	err = sc.Register("TestService", serv2)
	if err != nil {
		t.Fatal(err)
	}

	sc.Iter(func(name string, service common.Service) error {
		t.Logf("%s: %p", name, service)
		return nil
	})

	sc.ReverseIter(func(name string, service common.Service) error {
		t.Logf("%s: %p", name, service)
		return nil
	})
	var i int
	sc.IterWithPredicate(func(name string, service common.Service) error {
		t.Logf("IterWithPredicate ==>%s: %p", name, service)
		i++
		return nil
	}, func(name string) bool {
		return name != "TestService"
	})

	if i != 1 {
		t.Fatal("invalid IterWithPredicate ", i)
	}

	err = sc.UnRegister(LedgerService)
	if err != nil {
		t.Fatal(err)
	}

	sc.Iter(func(name string, service common.Service) error {
		t.Logf("%s: %p", name, service)
		return nil
	})

	if _, err := sc.Get(LedgerService); err == nil {
		t.Fatal("shouldn't find ledger service")
	}
}

func TestNewChainContext(t *testing.T) {
	cfgFile1 := filepath.Join(config.TestDataDir(), "config1", config.CfgFileName)
	cfgFile2 := filepath.Join(config.TestDataDir(), "config2", "test.json")
	t.Log(filepath.Dir(cfgFile2), filepath.Base(cfgFile2))
	cm := config.NewCfgManagerWithName(filepath.Dir(cfgFile2), filepath.Base(cfgFile2))
	cfg, err := cm.Config()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DataDir = filepath.Dir(cfgFile2)
	err = cm.Save()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(cfgFile1)
		_ = os.Remove(cfgFile2)
	}()

	c1 := NewServiceContext(cfgFile1)

	if f := c1.cfgFile; cfgFile1 != f {
		t.Fatalf("invalid config, exp: %s, act: %s", cfgFile1, f)
	}

	c2 := NewServiceContext(cfgFile2)
	if c1 == nil || c2 == nil {
		t.Fatal("failed to create context")
	} else {
		if c1.Id() == c2.Id() {
			t.Fatal("invalid c1 and c2")
		} else {
			t.Log(c1.Id(), c2.Id())
		}
	}

	c3 := NewServiceContext(cfgFile1)
	if c1 != c3 {
		t.Fatalf("invalid instance expect: %p,act :%p", c1, c3)
	}

	cfg1, err := c1.Config()
	if err != nil {
		t.Fatal(err)
	}
	cfg2, err := c2.Config()
	if err != nil {
		t.Fatal(err)
	}

	if cfg1.DataDir != filepath.Dir(cfgFile1) {
		t.Fatalf(cfg1.DataDir, filepath.Dir(cfgFile1))
	}
	if cfg2.DataDir != filepath.Dir(cfgFile2) {
		t.Fatalf(cfg2.DataDir, filepath.Dir(cfgFile2))
	}

	eb1 := c1.EventBus()
	eb2 := c2.EventBus()
	eb3 := c3.EventBus()
	if eb1 == eb2 {
		t.Fatal("eb1 shouldn't same as eb2")
	}

	if eb1 != eb3 {
		t.Fatal("eb1 shouldn same as eb3")
	}

	feb1 := c1.FeedEventBus()
	feb2 := c2.FeedEventBus()
	feb3 := c3.FeedEventBus()
	if feb1 == feb2 {
		t.Fatal("eb1 shouldn't same as eb2")
	}

	if feb1 != feb3 {
		t.Fatal("eb1 shouldn same as eb3")
	}
}

func TestChainContext_WaitForever(t *testing.T) {
	cfgFile := filepath.Join(config.TestDataDir(), "context", uuid.New().String(), config.CfgFileName)
	defer func() { _ = os.RemoveAll(cfgFile) }()

	ctx := NewServiceContext(cfgFile)
	err := ctx.Init(func() error {
		return ctx.Register("waitService", &waitService{})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = ctx.Start()
	if err != nil {
		t.Fatal(err)
	}

	ctx.WaitForever()

	if s, err := ctx.Service("waitService"); err == nil {
		if s.Status() != int32(common.Started) {
			t.Fatal("start failed", s.Status())
		}
	} else {
		t.Fatal(err)
	}
}

func TestSerivceContext_ConfigManager(t *testing.T) {
	cfgFile := filepath.Join(config.TestDataDir(), "config2", "test.json")
	ctx := NewServiceContext(cfgFile)
	err := ctx.Init(func() error {
		return ctx.Register("waitService", &waitService{})
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg, err := ctx.ConfigManager(func(cm *config.CfgManager) error {
		if cfg, err := cm.Load(); err != nil {
			return err
		} else {
			cfg.LogLevel = "info"
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("1, ctx==>%p,cfg==>%p", ctx, cfg)
	}
	ctx2 := NewServiceContext(cfgFile)
	if cfg2, err := ctx2.Config(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("2, ctx==>%p,cfg==>%p", ctx2, cfg2)
		if cfg2.LogLevel != "info" {
			t.Fatal("invalid loglevel")
		}
	}

}

func TestChainContext(t *testing.T) {
	cfgFile := filepath.Join(config.TestDataDir(), uuid.New().String(), "test.json")
	ctx := NewServiceContext(cfgFile)
	const name = "waitService"
	if err := ctx.Init(func() error {
		return ctx.Register(name, &waitService{})
	}); err != nil {
		t.Fatal(err)
	}

	if err := ctx.Start(); err != nil {
		t.Fatal(err)
	}

	if err := ctx.ReloadService(name); err != nil {
		t.Fatal(err)
	}

	if b := ctx.HasService(name); !b {
		t.Fatal()
	}

	if services, err := ctx.AllServices(); err != nil {
		t.Fatal(err)
	} else {
		if len(services) != 1 {
			t.Fatal()
		}
	}

	if err := ctx.Destroy(); err != nil {
		t.Fatal(err)
	}

	t.Log(ctx.Status())
}

func mockAccount() *sdk.Account {
	seed, _ := sdk.NewSeed()
	_, priv, _ := sdk.KeypairFromSeed(seed.String(), 0)
	return sdk.NewAccount(priv)
}
