package event

import (
	"sync"

	ct "github.com/qlcchain/qlc-hub/common/topic"
)

type FeedEventBus struct {
	sync.RWMutex
	feeds map[ct.TopicType]*Feed
}

func NewFeedEventBus() *FeedEventBus {
	return &FeedEventBus{
		feeds: make(map[ct.TopicType]*Feed),
	}
}

func (eb *FeedEventBus) Subscribe(topic ct.TopicType, ch interface{}) FeedSubscription {
	eb.Lock()
	defer eb.Unlock()

	f := eb.feeds[topic]
	if f == nil {
		f = &Feed{}
	}
	eb.feeds[topic] = f

	sub := f.Subscribe(ch)
	if sub == nil {
		return nil
	}

	return sub
}

func (eb *FeedEventBus) Unsubscribe(sub FeedSubscription) {
	sub.Unsubscribe()
}

func (eb *FeedEventBus) Publish(topic ct.TopicType, msg interface{}) {
	f := eb.findFeed(topic)
	if f == nil {
		return
	}

	f.Send(msg)
}

func (eb *FeedEventBus) LookupFeed(topic ct.TopicType) *Feed {
	eb.Lock()
	f := eb.feeds[topic]
	if f == nil {
		f = &Feed{}
	}
	eb.feeds[topic] = f
	eb.Unlock()

	return f
}

func (eb *FeedEventBus) findFeed(topic ct.TopicType) *Feed {
	eb.RLock()
	f := eb.feeds[topic]
	eb.RUnlock()

	return f
}
