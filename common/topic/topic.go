/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package topic

type TopicType string

//Topic type
const (
	EventRpcSyncCall TopicType = "rpcSyncCall"
)

// Sync state
type SyncState uint

const (
	SyncNotStart SyncState = iota
	Syncing
	SyncDone
	SyncFinish
)

var syncStatus = [...]string{
	SyncNotStart: "SyncNotStart",
	Syncing:      "Synchronizing",
	SyncDone:     "SyncDone",
	SyncFinish:   "SyncFinish",
}

func (s SyncState) String() string {
	if s > SyncFinish {
		return "unknown sync state"
	}
	return syncStatus[s]
}

func (s SyncState) IsSyncExited() bool {
	if s == SyncDone {
		return true
	}

	return false
}
