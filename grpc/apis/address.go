package apis

import (
	"sync"
	"time"

	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/pkg/log"
)

type AddressPool struct {
	addresses []qlctypes.Address
	lock      sync.RWMutex
	logger    *zap.SugaredLogger
}

func NewAddressPool() *AddressPool {
	return &AddressPool{addresses: []qlctypes.Address{}, logger: log.NewLogger("addresses")}
}

func AddressPools(address qlctypes.Address) *AddressPool {
	pool := NewAddressPool()
	pool.Enqueue(address)
	return pool
}

func (pool *AddressPool) Enqueue(address qlctypes.Address) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.addresses = append(pool.addresses, address)
	pool.logger.Infof("set account to pool: %s, pool length: %d", address, len(pool.addresses))
}

func (pool *AddressPool) Dequeue() qlctypes.Address {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if len(pool.addresses) > 0 {
		item := pool.addresses[0]
		pool.addresses = pool.addresses[1:]
		pool.logger.Infof("get account from pool: %s, pool length: %d", item, len(pool.addresses))
		return item
	} else {
		return qlctypes.ZeroAddress
	}
}

func (pool *AddressPool) DequeueSync() qlctypes.Address {
	t := time.After(30 * time.Second)
	for {
		search := pool.Dequeue()
		if search != qlctypes.ZeroAddress {
			return search
		}

		select {
		case <-t:
			return qlctypes.ZeroAddress
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (pool *AddressPool) Front() qlctypes.Address {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	item := pool.addresses[0]
	return item
}

func (pool *AddressPool) Search(address qlctypes.Address) qlctypes.Address {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for i, add := range pool.addresses {
		if add == address {
			pool.addresses = append(pool.addresses[:i], pool.addresses[i+1:]...)
			pool.logger.Infof("search account from pool: %s, pool length:  %d", add, len(pool.addresses))
			return add
		}
	}
	return qlctypes.ZeroAddress
}

func (pool *AddressPool) IsEmpty() bool {
	return len(pool.addresses) == 0
}

func (pool *AddressPool) Size() int {
	return len(pool.addresses)
}

func (pool *AddressPool) SearchSync(address qlctypes.Address) qlctypes.Address {
	t := time.After(60 * time.Second)
	for {
		search := pool.Search(address)
		if search != qlctypes.ZeroAddress {
			return search
		}

		select {
		case <-t:
			return qlctypes.ZeroAddress
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
