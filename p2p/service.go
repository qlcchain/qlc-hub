package p2p

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/qlc-hub/common/event"
	ctx "github.com/qlcchain/qlc-hub/services/context"
)

// service for qlc hub p2p network
type P2pService struct {
	subscriber *event.ActorSubscriber
	node       *Node
	dispatcher *Dispatcher
	msgEvent   event.EventBus
	msgService *MessageService
	cc         *ctx.ServiceContext
}

// NewQlcService create netService
func NewP2pService(cfgFile string) (*P2pService, error) {
	cc := ctx.NewServiceContext(cfgFile)
	cfg, _ := cc.Config()
	node, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	ps := &P2pService{
		node:       node,
		dispatcher: NewDispatcher(),
		msgEvent:   cc.EventBus(),
		cc:         cc,
	}
	node.SetP2pService(ps)
	msgService := NewMessageService(ps)
	ps.msgService = msgService
	return ps, nil
}

// Node return the peer node
func (ps *P2pService) Node() *Node {
	return ps.node
}

// EventQueue return EventQueue
func (ps *P2pService) MessageEvent() event.EventBus {
	return ps.msgEvent
}

// Start start p2p manager.
func (ps *P2pService) Start() error {
	//ns.node.logger.VInfo("Starting QlcService...")

	// start dispatcher.
	ps.dispatcher.Start()

	//set event
	if err := ps.setEvent(); err != nil {
		return err
	}

	// start node.
	if err := ps.node.StartServices(); err != nil {
		ps.dispatcher.Stop()
		ps.node.logger.Error("Failed to start QlcService.")
		return err
	}
	// start msgService
	ps.msgService.Start()
	ps.node.logger.Info("started p2p service.")
	return nil
}

func (ps *P2pService) setEvent() error {
	ps.subscriber = event.NewActorSubscriber(event.SpawnWithPool(func(c actor.Context) {
		//switch msg := c.Message().(type) {
	}), ps.msgEvent)

	if err := ps.subscriber.Subscribe(); err != nil {
		ps.node.logger.Error(err)
		return err
	}

	return nil
}

// Stop stop p2p manager.
func (ps *P2pService) Stop() error {
	// ns.node.logger.VInfo("Stopping QlcService...")

	// this must be the first step
	err := ps.subscriber.UnsubscribeAll()
	if err != nil {
		return err
	}

	if err := ps.node.Stop(); err != nil {
		return err
	}

	ps.dispatcher.Stop()
	ps.msgService.Stop()

	time.Sleep(100 * time.Millisecond)
	return nil
}

// Register register the subscribers.
func (ps *P2pService) Register(subscribers ...*Subscriber) {
	ps.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ps *P2pService) Deregister(subscribers *Subscriber) {
	ps.dispatcher.Deregister(subscribers)
}

// PutMessage put snyc message to dispatcher.
func (ps *P2pService) PutSyncMessage(msg *Message) {
	ps.dispatcher.PutSyncMessage(msg)
}

// PutMessage put dpos message to dispatcher.
func (ps *P2pService) PutMessage(msg *Message) {
	ps.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ps *P2pService) Broadcast(name MessageType, value interface{}) {
	ps.node.BroadcastMessage(name, value)
}

// SendMessageToPeer send message to a peer.
func (ps *P2pService) SendMessageToPeer(messageName MessageType, value interface{}, peerID string) error {
	return ps.node.SendMessageToPeer(messageName, value, peerID)
}
