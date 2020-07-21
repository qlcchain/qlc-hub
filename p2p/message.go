package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"

	"github.com/qlcchain/qlc-go-sdk/pkg/types"
)

type MessageType byte

const (
	MessageHeaderLength       = 13
	MessageMagicNumberEndIdx  = 3
	MessageVersionEndIdx      = 4
	MessageTypeEndIdx         = 5
	MessageDataLengthEndIdx   = 9
	MessageDataCheckSumEndIdx = 13
)

// Error types
var (
	MagicNumber                   = []byte{0x51, 0x4C, 0x43} //QLC
	ErrInvalidMessageHeaderLength = errors.New("invalid message header length")
	ErrInvalidMessageDataLength   = errors.New("invalid message data length")
	ErrInvalidMagicNumber         = errors.New("invalid magic number")
	ErrInvalidDataCheckSum        = errors.New("invalid data checksum")
)

type HubMessage struct {
	content     []byte
	messageType MessageType
}

// MagicNumber return magicNumber
func (message *HubMessage) MagicNumber() []byte {
	return message.content[:MessageMagicNumberEndIdx]
}

func (message *HubMessage) Version() byte {
	return message.content[MessageMagicNumberEndIdx]
}

func (message *HubMessage) MessageType() MessageType {
	return MessageType(message.content[MessageVersionEndIdx])
}

func (message *HubMessage) MessageData() []byte {
	return message.content[MessageDataCheckSumEndIdx:]
}

// DataLength return dataLength
func (message *HubMessage) DataLength() uint32 {
	return Uint32(message.content[MessageTypeEndIdx:MessageDataLengthEndIdx])
}

// DataCheckSum return data checkSum
func (message *HubMessage) DataCheckSum() uint32 {
	return Uint32(message.content[MessageDataLengthEndIdx:MessageDataCheckSumEndIdx])
}

// HeaderData return HeaderData
func (message *HubMessage) HeaderData() []byte {
	return message.content[:MessageDataLengthEndIdx]
}

// NewHubMessage new qlc message
func NewHubMessage(data []byte, currentVersion byte, messageType MessageType) []byte {
	message := &HubMessage{
		content: make([]byte, MessageHeaderLength+len(data)),
	}
	// copy header.
	copy(message.content[0:MessageMagicNumberEndIdx], MagicNumber)
	message.content[MessageMagicNumberEndIdx] = currentVersion
	message.content[MessageVersionEndIdx] = byte(messageType)

	//copy datalength
	copy(message.content[MessageTypeEndIdx:MessageDataLengthEndIdx], FromUint32(uint32(len(data))))

	// copy data.
	copy(message.content[MessageDataCheckSumEndIdx:], data)

	// data checksum.
	dataCheckSum := crc32.ChecksumIEEE(message.content[MessageDataCheckSumEndIdx:])
	copy(message.content[MessageDataLengthEndIdx:MessageDataCheckSumEndIdx], FromUint32(uint32(dataCheckSum)))

	return message.content
}

// ParseqlcMessage parse qlc message
func ParseHubMessage(data []byte) (*HubMessage, error) {
	if len(data) < MessageHeaderLength {
		return nil, ErrInvalidMessageHeaderLength
	}
	message := &HubMessage{
		content: make([]byte, MessageHeaderLength),
	}
	copy(message.content, data)
	if err := message.VerifyHeader(); err != nil {
		return nil, err
	}
	message.messageType = message.MessageType()
	return message, nil
}

// ParseMessageData parse qlc message data
func (message *HubMessage) ParseMessageData(data []byte) error {
	if uint32(len(data)) < message.DataLength() {
		return ErrInvalidMessageDataLength
	}
	message.content = append(message.content, data[:message.DataLength()]...)
	return message.VerifyData()
}

//VerifyHeader verify qlc message header
func (message *HubMessage) VerifyHeader() error {
	if !Equal(MagicNumber, message.MagicNumber()) {
		return ErrInvalidMagicNumber
	}
	return nil
}

// VerifyData verify qlc message data
func (message *HubMessage) VerifyData() error {
	dataCheckSum := crc32.ChecksumIEEE(message.MessageData())
	if dataCheckSum != message.DataCheckSum() {
		return ErrInvalidDataCheckSum
	}
	return nil
}

// FromUint32 decodes uint32.
func FromUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// Uint32 encodes []byte.
func Uint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// Equal checks whether byte slice a and b are equal.
func Equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Subscriber subscriber.
type Subscriber struct {
	// msgChan chan for subscribed message.
	msgChan chan *Message

	// msgType message type to subscribe
	msgType MessageType
}

// NewSubscriber return new Subscriber instance.
func NewSubscriber(msgChan chan *Message, msgType MessageType) *Subscriber {
	return &Subscriber{msgChan, msgType}
}

// MessageType return msgTypes.
func (s *Subscriber) MessageType() MessageType {
	return s.msgType
}

// MessageChan return msgChan.
func (s *Subscriber) MessageChan() chan *Message {
	return s.msgChan
}

// Message struct
type Message struct {
	messageType MessageType
	from        string
	data        []byte //removed the header
	content     []byte //complete message data
}

// NewBaseMessage new base message
func NewMessage(messageType MessageType, from string, data []byte, content []byte) *Message {
	return &Message{messageType: messageType, from: from, data: data, content: content}
}

// MessageType get message type
func (msg *Message) MessageType() MessageType {
	return msg.messageType
}

// MessageFrom get message who send
func (msg *Message) MessageFrom() string {
	return msg.from
}

// Data get the message data
func (msg *Message) Data() []byte {
	return msg.data
}

// Content get the message content
func (msg *Message) Content() []byte {
	return msg.content
}

// Hash return the message hash
func (msg *Message) Hash() types.Hash {
	hash, _ := types.HashBytes(msg.content)
	return hash
}

// String get the message to string
func (msg *Message) String() string {
	return fmt.Sprintf("Message {type:%d; data:%s; from:%s}",
		msg.messageType,
		msg.data,
		msg.from,
	)
}
