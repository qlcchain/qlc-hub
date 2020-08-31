package store

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
)

var (
	lcache = make(map[string]*Store)
	lock   = sync.RWMutex{}
)

type Store struct {
	io.Closer
	dir    string
	store  db.Store
	logger *zap.SugaredLogger
}

func getDefaultDir(dir string) string {
	if dir != "" {
		return dir
	} else {
		//todo set default path
		return config.DefaultDataDir()
	}
}

func NewStore(dir string) (*Store, error) {
	lock.Lock()
	defer lock.Unlock()
	dir = getDefaultDir(dir)

	if _, ok := lcache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			return nil, fmt.Errorf("NewBadgerStore: %s", err)
		}
		l := &Store{
			dir:    dir,
			store:  store,
			logger: log.NewLogger("store"),
		}
		lcache[dir] = l
	}
	return lcache[dir], nil
}

//CloseLedger force release all store instance
func CloseLedger() {
	for k, v := range lcache {
		if v != nil {
			v.Close()
		}
		lock.Lock()
		delete(lcache, k)
		lock.Unlock()
	}
}

func (l *Store) Close() error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := lcache[l.dir]; ok {
		if err := l.store.Close(); err != nil {
			return err
		}
		l.logger.Info("badger closed")
		delete(lcache, l.dir)
		return nil
	}
	return nil
}

type KeyPrefix byte

const (
	KeyPrefixLockerInfo KeyPrefix = iota
)

var (
	ErrLockerInfoExists   = errors.New("locker info already exists")
	ErrLockerInfoNotFound = errors.New("locker info not found")
)
