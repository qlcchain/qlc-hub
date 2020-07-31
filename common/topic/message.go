package topic

// MessageType a string for message type.
type MessageType byte

type EventRPCSyncCallMsg struct {
	Name string
	In   interface{}
	Out  interface{}

	ResponseChan chan interface{}
}
