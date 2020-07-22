package topic

import "github.com/qlcchain/qlc-hub/common/types"

// MessageType a string for message type.
type MessageType byte

type EventAddP2PStreamMsg struct {
	PeerID   string
	PeerInfo string
}

type EventDeleteP2PStreamMsg struct {
	PeerID string
}

type EventP2PConnectPeersMsg struct {
	PeersInfo []*types.PeerInfo
}

type EventP2POnlinePeersMsg struct {
	PeersInfo []*types.PeerInfo
}

type EventBroadcastMsg struct {
	Type    MessageType
	Message interface{}
}
