package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/qlcchain/qlc-hub/common/topic"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	ping "github.com/qlcchain/qlc-hub/p2p/pinger"
)

// Stream Errors
var (
	ErrStreamIsNotConnected = errors.New("stream is not connected")
	ErrNoStream             = errors.New("no stream")
	ErrCloseStream          = errors.New("stream close error")
)

// Stream define the structure of a stream in p2p network
type Stream struct {
	syncMutex         sync.Mutex
	pid               peer.ID
	addr              ma.Multiaddr
	stream            network.Stream
	node              *Node
	quitWriteCh       chan bool
	messageNotifyChan chan int
	messageChan       chan *HubMessage
	rtt               time.Duration
	pingCtx           context.Context
	pingCancel        context.CancelFunc
	pingTimeoutTimes  int
	pingResult        <-chan ping.Result
	globalVersion     string
	p2pVersion        byte
	lastUpdateTime    string
}

// NewStream return a new Stream
func NewStream(stream network.Stream, node *Node) *Stream {
	return newStreamInstance(stream.Conn().RemotePeer(), stream.Conn().RemoteMultiaddr(), stream, node)
}

// NewStreamFromPID return a new Stream based on the pid
func NewStreamFromPID(pid peer.ID, node *Node) *Stream {
	return newStreamInstance(pid, nil, nil, node)
}

func newStreamInstance(pid peer.ID, addr ma.Multiaddr, stream network.Stream, node *Node) *Stream {
	ctx, cancel := context.WithCancel(context.Background())
	return &Stream{
		pid:               pid,
		addr:              addr,
		stream:            stream,
		node:              node,
		quitWriteCh:       make(chan bool, 1),
		messageNotifyChan: make(chan int, 60*1024),
		messageChan:       make(chan *HubMessage, 20*1024),
		pingCtx:           ctx,
		pingCancel:        cancel,
		pingTimeoutTimes:  0,
	}
}

// Connect to the stream
func (s *Stream) Connect() error {
	//s.node.logger.VInfo("Connecting to peer.")

	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.ctx,
		s.pid,
		ProtocolID,
	)
	if err != nil {
		return err
	}
	s.stream = stream
	s.addr = stream.Conn().RemoteMultiaddr()
	return nil
}

// IsConnected return if the stream is connected
func (s *Stream) IsConnected() bool {
	return s.stream != nil
}

func (s *Stream) String() string {
	addrStr := ""
	if s.addr != nil {
		addrStr = s.addr.String()
	}

	return fmt.Sprintf("Peer Stream: %s,%s", s.pid.Pretty(), addrStr)
}

// StartLoop start stream ping loop.
func (s *Stream) StartLoop() {
	go s.writeLoop()
	go s.readLoop()
}

func (s *Stream) readLoop() {
	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			//			s.node.logger.Error(err)
			err = s.close()
			if err != nil {
				s.node.logger.Error(err)
			}
			return
		}
	}
	s.node.logger.Info("connect ", s.pid.Pretty(), " success")

	s.node.netService.MessageEvent().Publish(topic.EventAddP2PStream, &topic.EventAddP2PStreamMsg{PeerID: s.pid.Pretty(), PeerInfo: s.addr.String()})

	// loop.
	buf := make([]byte, 1024*4)
	messageBuffer := make([]byte, 0)

	var message *HubMessage

	for {
		n, err := s.stream.Read(buf)
		if err != nil {
			s.node.logger.Debugf("Error occurred when reading data from network connection.")
			if err := s.close(); err != nil {
				s.node.logger.Error(err)
			}
			return
		}

		messageBuffer = append(messageBuffer, buf[:n]...)

		for {
			if message == nil {
				var err error

				// waiting for header data.
				if len(messageBuffer) < MessageHeaderLength {
					// continue reading.
					break
				}
				message, err = ParseHubMessage(messageBuffer)
				if err != nil {
					return
				}
				messageBuffer = messageBuffer[MessageHeaderLength:]
			}
			// waiting for data.
			if len(messageBuffer) < int(message.DataLength()) {
				// continue reading.
				break
			}
			if err := message.ParseMessageData(messageBuffer); err != nil {
				return
			}
			// remove data from buffer.
			messageBuffer = messageBuffer[message.DataLength():]

			// handle message.
			s.handleMessage(message)
			// reset message.
			message = nil
		}
	}
}

func (s *Stream) writeLoop() {
	for {
		select {
		case <-s.quitWriteCh:
			s.node.logger.Debug("Quiting Stream Write Loop.")
			return
		case <-s.messageNotifyChan:
			select {
			case message := <-s.messageChan:
				_ = s.WriteQlcMessage(message)
				continue
			default:
			}
		}
	}
}

// Close close the stream
func (s *Stream) close() error {
	// Add lock & close flag to prevent multi call.
	//s.syncMutex.Lock()
	//defer s.syncMutex.Unlock()
	//s.node.logger.VInfo("Closing stream.")

	if s.stream != nil {
		s.node.netService.MessageEvent().Publish(topic.EventDeleteP2PStream,
			&topic.EventDeleteP2PStreamMsg{PeerID: s.pid.Pretty()})
	}

	// cleanup.
	s.node.streamManager.RemoveStream(s)

	// quit.
	s.quitWriteCh <- true
	s.pingCancel()
	// close stream.
	if s.stream != nil {
		if err := s.stream.Close(); err != nil {
			return ErrCloseStream
		}
	}
	return nil
}

// SendMessage send msg to peer
func (s *Stream) SendMessageToPeer(messageType MessageType, data []byte) error {
	version := p2pVersion
	message := NewHubMessage(data, byte(version), messageType)
	qlcMessage := &HubMessage{
		messageType: messageType,
		content:     message,
	}

	err := s.SendMessageToChan(qlcMessage)

	return err
}

// WriteQlcMessage write qlc msg in the stream
func (s *Stream) WriteQlcMessage(message *HubMessage) error {
	err := s.Write(message.content)

	return err
}

func (s *Stream) Write(data []byte) error {
	if s.stream == nil {
		if err := s.close(); err != nil {
			return ErrCloseStream
		}

		return ErrStreamIsNotConnected
	}

	// at least 5kb/s to write message
	//deadline := time.Now().Add(time.Duration(len(data)/1024/5+1) * time.Second)
	//if err := s.stream.SetWriteDeadline(deadline); err != nil {
	//	return err
	//}

	n, err := s.stream.Write(data)
	if err != nil {
		s.node.logger.Debugf("Failed to send message to peer [%s].", s.pid.Pretty())
		//s.close()
		return err
	}
	s.node.logger.Debugf("%d byte send to %v ", n, s.pid.Pretty())
	return nil
}

func (s *Stream) handleMessage(message *HubMessage) {
	s.p2pVersion = message.Version()
	if message.Version() < byte(p2pVersion) {
		s.node.logger.Debugf("message Version [%d] is less then p2pVersion [%d]", message.Version(), p2pVersion)
		return
	}
	m := NewMessage(message.MessageType(), s.pid.Pretty(), message.MessageData(), message.content)
	s.node.netService.PutSyncMessage(m)
}

// SendMessage send msg to buffer
func (s *Stream) SendMessageToChan(message *HubMessage) error {
	select {
	case s.messageChan <- message:
	default:
		s.node.logger.Debugf("Received too many normal priority message.")
		return nil
	}
	select {
	case s.messageNotifyChan <- 1:
	default:
		s.node.logger.Debugf("Received too many message notifyChan.")
		return nil
	}
	return nil
}
