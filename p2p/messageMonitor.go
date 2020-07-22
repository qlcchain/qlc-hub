package p2p

import (
	"context"
	"errors"
)

const (
	MonitorMsgChanSize = 65535
)

//  Message Type
const (
	TestMessage MessageType = iota
)

type MessageService struct {
	netService *P2pService
	ctx        context.Context
	cancel     context.CancelFunc
	messageCh  chan *Message
}

// NewService return new Service.
func NewMessageService(netService *P2pService) *MessageService {
	ctx, cancel := context.WithCancel(context.Background())
	ms := &MessageService{
		ctx:        ctx,
		cancel:     cancel,
		messageCh:  make(chan *Message, MonitorMsgChanSize),
		netService: netService,
	}
	return ms
}

// Start start message service.
func (ms *MessageService) Start() {
	// register the network handler.
	netService := ms.netService
	netService.Register(NewSubscriber(ms.messageCh, TestMessage))

	// start loop().
	go ms.startLoop()
}

func (ms *MessageService) startLoop() {
	ms.netService.node.logger.Info("Started Message Service.")
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.messageCh:
			switch message.MessageType() {
			case TestMessage:
				go ms.onTestReq(message)
			default:
				ms.netService.node.logger.Error("Received unknown message.")
			}
		}
	}
}

func (ms *MessageService) onTestReq(message *Message) {
	ms.netService.node.logger.Info(string(message.data))
}

func (ms *MessageService) Stop() {
	//ms.netService.node.logger.VInfo("stopped message monitor")
	// quit.
	ms.cancel()
	ms.netService.Deregister(NewSubscriber(ms.messageCh, TestMessage))
}

func marshalMessage(messageName MessageType, value interface{}) ([]byte, error) {
	switch messageName {
	case TestMessage:
		t := "this is a test message"
		return []byte(t), nil
	default:
		return nil, errors.New("unKnown Message Type")
	}
}
