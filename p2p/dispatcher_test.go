package p2p

import (
	"testing"
)

func TestDispatcher(t *testing.T) {
	dp := NewDispatcher()
	sb := NewSubscriber(make(chan *Message, 128), TestMessage)
	types := sb.MessageType()
	dp.Register(sb)
	mt, _ := dp.subscribersMap.Load(types)
	if mt == nil {
		t.Fatal("register fail")
	}
	dp.Deregister(sb)
	_, ok := dp.subscribersMap.Load(types)

	if ok {
		t.Fatal("deregister fail")
	}
}
