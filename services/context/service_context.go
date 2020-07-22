/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package context

import (
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	sdk "github.com/qlcchain/qlc-go-sdk/pkg/types"

	"github.com/qlcchain/qlc-hub/common"
	"github.com/qlcchain/qlc-hub/common/event"
	"github.com/qlcchain/qlc-hub/common/hashmap"
	"github.com/qlcchain/qlc-hub/common/topic"
	"github.com/qlcchain/qlc-hub/common/types"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/log"
)

var cache = hashmap.New(10)

var ErrPoVNotFinish = errors.New("pov sync is not finished, please check it")

const (
	LedgerService       = "ledgerService"
	BootNodeHttpService = "bootNodeHttpService"
	P2PService          = "P2PService"
	LogService          = "logService"
)

type serviceManager interface {
	common.Service
	Register(name string, service common.Service) error
	UnRegister(name string) error
	AllServices() ([]common.Service, error)
	Service(name string) (common.Service, error)
	HasService(name string) bool
	//Control
	ReloadService(name string) error
	RestartAll() error
	// config
	ConfigManager() (*config.CfgManager, error)
	Config() (*config.Config, error)
	EventBus() event.EventBus
}

type Option func(cm *config.CfgManager) error

func NewServiceContext(cfgFile string) *ServiceContext {
	var dataDir string
	if len(cfgFile) == 0 {
		dataDir = config.DefaultDataDir()
		cfgFile = path.Join(dataDir, config.CfgFileName)
	} else {
		cm := config.NewCfgManagerWithFile(cfgFile)
		dataDir, _ = cm.ParseDataDir()
	}
	id := sdk.HashData([]byte(dataDir)).String()
	if v, ok := cache.GetStringKey(id); ok {
		return v.(*ServiceContext)
	} else {
		sr := &ServiceContext{
			services:         newServiceContainer(),
			cfgFile:          cfgFile,
			chainID:          id,
			connectPeersPool: new(sync.Map),
		}
		cache.Set(id, sr)
		return sr
	}
}

type ServiceContext struct {
	common.ServiceLifecycle
	services         *serviceContainer
	cm               *config.CfgManager
	cfgFile          string
	chainID          string
	locker           sync.RWMutex
	accounts         []*sdk.Account
	subscriber       *event.ActorSubscriber
	connectPeersPool *sync.Map
	connectPeersInfo []*types.PeerInfo
	onlinePeersInfo  []*types.PeerInfo
}

func (sc *ServiceContext) EventBus() event.EventBus {
	return event.GetEventBus(sc.Id())
}

func (sc *ServiceContext) FeedEventBus() *event.FeedEventBus {
	return event.GetFeedEventBus(sc.Id())
}

func (sc *ServiceContext) GetPeersPool() map[string]string {
	p := make(map[string]string)
	sc.connectPeersPool.Range(func(key, value interface{}) bool {
		peerId := key.(string)
		addr := value.(string)
		p[peerId] = addr
		return true
	})
	return p
}

func (sc *ServiceContext) GetConnectPeersInfo() []*types.PeerInfo {
	return sc.connectPeersInfo
}

func (sc *ServiceContext) GetOnlinePeersInfo() []*types.PeerInfo {
	return sc.onlinePeersInfo
}

func (sc *ServiceContext) ConfigFile() string {
	return sc.cfgFile
}

func (sc *ServiceContext) Init(fn func() error) error {
	if !sc.PreInit() {
		return errors.New("pre init fail")
	}
	defer sc.PostInit()

	if fn != nil {
		err := fn()
		if err != nil {
			return err
		}
	}

	err := sc.services.IterWithPredicate(func(name string, service common.Service) error {
		err := service.Init()
		if err != nil {
			return err
		}
		log.Root.Infof("%s init successfully", name)
		return nil
	}, func(name string) bool {
		return name != LogService
	})
	if err != nil {
		return err
	}

	sc.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *topic.EventAddP2PStreamMsg:
			if _, ok := sc.connectPeersPool.Load(msg.PeerID); ok {
				sc.connectPeersPool.Delete(msg.PeerID)
			}
			sc.connectPeersPool.Store(msg.PeerID, msg.PeerInfo)
		case *topic.EventDeleteP2PStreamMsg:
			if _, ok := sc.connectPeersPool.Load(msg.PeerID); ok {
				sc.connectPeersPool.Delete(msg.PeerID)
			}
		case *topic.EventP2PConnectPeersMsg:
			sc.connectPeersInfo = msg.PeersInfo
		case *topic.EventP2POnlinePeersMsg:
			sc.onlinePeersInfo = msg.PeersInfo
		}
	}), sc.EventBus())

	return sc.subscriber.Subscribe(topic.EventOnlinePeersInfo, topic.EventPeersInfo, topic.EventAddP2PStream, topic.EventDeleteP2PStream)
}

func (sc *ServiceContext) Start() error {
	if !sc.PreStart() {
		return errors.New("pre start fail")
	}
	defer sc.PostStart()

	sc.services.Iter(func(name string, service common.Service) error {
		err := service.Start()
		if err != nil {
			return fmt.Errorf("%s, %s", name, err)
		}
		log.Root.Infof("%s start successfully", name)
		return nil
	})

	return nil
}

func (sc *ServiceContext) Stop() error {
	if !sc.PreStop() {
		return errors.New("pre stop fail")
	}
	defer sc.PostStop()

	sc.services.ReverseIter(func(name string, service common.Service) error {
		err := service.Stop()
		if err != nil {
			return err
		}
		log.Root.Infof("%s stop successfully", name)
		return nil
	})

	if sc.subscriber != nil {
		return sc.subscriber.UnsubscribeAll()
	}

	return nil
}

func (sc *ServiceContext) Status() int32 {
	return sc.State()
}

func (sc *ServiceContext) SetAccounts(accounts []*sdk.Account) {
	sc.locker.Lock()
	defer sc.locker.Unlock()
	sc.accounts = accounts
}

func (sc *ServiceContext) Accounts() []*sdk.Account {
	sc.locker.RLock()
	defer sc.locker.RUnlock()
	return sc.accounts
}

func (sc *ServiceContext) Id() string {
	return sc.chainID
}

func (sc *ServiceContext) Register(name string, service common.Service) error {
	return sc.services.Register(name, service)
}

func (sc *ServiceContext) HasService(name string) bool {
	return sc.services.HasService(name)
}

func (sc *ServiceContext) UnRegister(name string) error {
	return sc.services.UnRegister(name)
}

func (sc *ServiceContext) AllServices() ([]common.Service, error) {
	var services []common.Service
	sc.services.Iter(func(name string, service common.Service) error {
		services = append(services, service)
		return nil
	})
	return services, nil
}

func (sc *ServiceContext) WaitForever() {
	count := len(sc.services.services)
	for {
		counter := 0
		sc.services.Iter(func(name string, service common.Service) error {
			if service.Status() == int32(common.Started) {
				counter++
			} else {
				fmt.Println(name, service.Status())
			}
			// return fmt.Errorf("%s, %d", name, service.Status())
			return nil
		})
		if counter == count {
			return
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
}

func (sc *ServiceContext) Service(name string) (common.Service, error) {
	return sc.services.Get(name)
}

func (sc *ServiceContext) ReloadService(name string) error {
	service, err := sc.Service(name)
	if err != nil {
		return err
	}

	return reloadService(service)
}

func (sc *ServiceContext) RestartAll() error {
	panic("implement me")
}

func (sc *ServiceContext) Destroy() error {
	err := sc.Stop()
	if err != nil {
		return err
	}

	id := sc.Id()
	if _, ok := cache.GetStringKey(id); ok {
		cache.Del(id)
	}

	return nil
}

func (sc *ServiceContext) ConfigManager(opts ...Option) (*config.CfgManager, error) {
	sc.locker.Lock()
	defer sc.locker.Unlock()
	if sc.cm == nil {
		sc.cm = config.NewCfgManagerWithFile(sc.cfgFile)
		_, err := sc.cm.Load()
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		_ = opt(sc.cm)
	}

	return sc.cm, nil
}

func (sc *ServiceContext) Config() (*config.Config, error) {
	cm, err := sc.ConfigManager()
	if err != nil {
		return nil, err
	}
	return cm.Config()
}

func reloadService(s common.Service) error {
	err := s.Stop()
	if err != nil {
		return err
	}

	err = s.Init()
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return err
	}
	return nil
}

type serviceContainer struct {
	locker   sync.RWMutex
	services map[string]common.Service
	names    []string
}

func newServiceContainer() *serviceContainer {
	return &serviceContainer{
		locker:   sync.RWMutex{},
		services: make(map[string]common.Service),
		names:    []string{},
	}
}

func (sc *serviceContainer) Register(name string, s common.Service) error {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	if _, ok := sc.services[name]; ok {
		return fmt.Errorf("service[%s] already exist", name)
	} else {
		sc.services[name] = s
		sc.names = append(sc.names, name)
		return nil
	}
}

func (sc *serviceContainer) UnRegister(name string) error {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	if v, ok := sc.services[name]; ok {
		_ = v.Stop()
		delete(sc.services, name)
		for idx, n := range sc.names {
			if n == name {
				sc.names = append(sc.names[:idx], sc.names[idx+1:]...)
				break
			}
		}
		return nil
	} else {
		return fmt.Errorf("service[%s] not exist", name)
	}
}

func (sc *serviceContainer) Get(name string) (common.Service, error) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	if v, ok := sc.services[name]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("service[%s] not exist", name)
	}
}

func (sc *serviceContainer) HasService(name string) bool {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	if _, ok := sc.services[name]; ok {
		return true
	}

	return false
}

func (sc *serviceContainer) Iter(fn func(name string, service common.Service) error) {
	_ = sc.IterWithPredicate(fn, func(name string) bool {
		return true
	})
}

func (sc *serviceContainer) IterWithPredicate(fn func(name string, service common.Service) error,
	predicate func(name string) bool) error {
	sc.locker.RLock()
	defer sc.locker.RUnlock()
	for idx := range sc.names {
		name := sc.names[idx]
		if service, ok := sc.services[name]; ok && predicate(name) {
			err := fn(name, service)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *serviceContainer) ReverseIter(fn func(name string, service common.Service) error) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	for i := len(sc.names) - 1; i >= 0; i-- {
		name := sc.names[i]
		if service, ok := sc.services[name]; ok {
			err := fn(name, service)
			if err != nil {
				break
			}
		}
	}
}
