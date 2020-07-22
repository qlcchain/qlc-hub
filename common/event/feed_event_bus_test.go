package event

import (
	"testing"
	"time"

	"github.com/qlcchain/qlc-hub/common/topic"
)

func TestFeedEventBus_PubSub1(t *testing.T) {
	feb := NewFeedEventBus()
	ch1 := make(chan *topic.EventAddP2PStreamMsg)
	ch2 := make(chan *topic.EventAddP2PStreamMsg)
	feb.Subscribe(topic.EventRpcSyncCall, ch1)
	feb.Subscribe(topic.EventRpcSyncCall, ch2)

	ch1RecvOk := false
	ch2RecvOk := false
	go func() {
		wt := time.NewTimer(time.Second)
		for {
			select {
			case <-ch1:
				ch1RecvOk = true
			case <-ch2:
				ch2RecvOk = true
			case <-wt.C:
				return
			}
		}
	}()

	feb.Publish(topic.EventRpcSyncCall, &topic.EventAddP2PStreamMsg{})

	time.Sleep(100 * time.Millisecond)

	if !ch1RecvOk {
		t.Errorf("ch1 does not recv msg")
	}
	if !ch2RecvOk {
		t.Errorf("ch2 does not recv msg")
	}
}
