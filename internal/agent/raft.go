package agent

import "github.com/hashicorp/raft"

type RaftStore interface {
	raft.StableStore
	raft.LogStore
	Close() error
}
