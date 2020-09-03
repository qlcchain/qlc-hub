package apis

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"go.uber.org/zap"
)

func sha256(r string) string {
	panic("implement me")
}

var maxRHashSzie = 1000
var timeout = 24 * time.Hour

var glock = gcache.New(maxRHashSzie).Expiration(timeout).LRU().Build()

//todo delete data
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
