package apis

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"go.uber.org/zap"
)

var maxRHashSize = 10240
var timeout = 24 * time.Hour

var glock = gcache.New(maxRHashSize).Expiration(timeout).LRU().Build()

func lock(rHash string, logger *zap.SugaredLogger) {
	if v, err := glock.Get(rHash); err != nil {
		mutex := &sync.Mutex{}
		if err := glock.Set(rHash, mutex); err != nil {
			logger.Errorf("set lock fail: %s [%s]", err, rHash)
		}
		mutex.Lock()
	} else {
		if l, ok := v.(*sync.Mutex); ok {
			l.Lock()
		} else {
			logger.Errorf("invalid lock type [%s]", rHash)
		}
	}
}

func unlock(rHash string, logger *zap.SugaredLogger) {
	if v, err := glock.Get(rHash); err != nil {
		logger.Errorf("can not get lock: %s [%s]", err, rHash)
	} else {
		if l, ok := v.(*sync.Mutex); ok {
			l.Unlock()
		} else {
			logger.Errorf("invalid lock type [%s]", rHash)
		}
	}
}
