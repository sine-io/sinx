package agent

import (
	"io"

	"github.com/hashicorp/raft"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"

	sxproto "github.com/sine-io/sinx/types"
)

// Raft finite state machine (FSM) is used to apply Raft log entries
// to the key-value store.

// MessageType is the type to encode FSM commands.
type MessageType uint8

const (
	// SetJobType is the command used to store a job in the store.
	SetJobType MessageType = iota
	// DeleteJobType is the command used to delete a Job from the store.
	DeleteJobType
	// SetExecutionType is the command used to store an Execution to the store.
	SetExecutionType
	// DeleteExecutionsType is the command used to delete executions from the store.
	DeleteExecutionsType
	// ExecutionDoneType is the command to perform the logic needed once an execution
	// is done.
	ExecutionDoneType
)

// LogApplier is the definition of a function that can apply a Raft log
type LogApplier func(buf []byte, index uint64) any

// LogAppliers is a mapping of the Raft MessageType to the appropriate log
// applier
type LogAppliers map[MessageType]LogApplier

type raftFSM struct {
	store Storage

	// proAppliers holds the set of pro only LogAppliers
	proAppliers LogAppliers
	logger      zerolog.Logger
}

// newRaftFSM is used to construct a new FSM with a blank state
func newRaftFSM(store Storage, logAppliers LogAppliers, logger zerolog.Logger) *raftFSM {
	return &raftFSM{
		store:       store,
		proAppliers: logAppliers,
		logger:      logger,
	}
}

// Apply applies a Raft log entry to the key-value store.
func (d *raftFSM) Apply(l *raft.Log) any {
	buf := l.Data
	msgType := MessageType(buf[0])

	d.logger.Debug().Any("command", msgType).Msg("fsm: received command")

	switch msgType {
	case SetJobType:
		return d.applySetJob(buf[1:])
	case DeleteJobType:
		return d.applyDeleteJob(buf[1:])
	case ExecutionDoneType:
		return d.applyExecutionDone(buf[1:])
	case SetExecutionType:
		return d.applySetExecution(buf[1:])
	}

	// Check enterprise only message types.
	if applier, ok := d.proAppliers[msgType]; ok {
		return applier(buf[1:], l.Index)
	}

	return nil
}

func (d *raftFSM) applySetJob(buf []byte) any {
	var pj sxproto.Job
	if err := proto.Unmarshal(buf, &pj); err != nil {
		return err
	}
	job := NewJobFromProto(&pj)
	if err := d.store.SetJob(job, false); err != nil {
		return err
	}
	return nil
}

func (d *raftFSM) applyDeleteJob(buf []byte) any {
	var djr sxproto.DeleteJobRequest
	if err := proto.Unmarshal(buf, &djr); err != nil {
		return err
	}
	job, err := d.store.DeleteJob(djr.GetJobName())
	if err != nil {
		return err
	}
	return job
}

func (d *raftFSM) applyExecutionDone(buf []byte) interface{} {
	var execDoneReq sxproto.ExecutionDoneRequest
	if err := proto.Unmarshal(buf, &execDoneReq); err != nil {
		return err
	}
	execution := NewExecutionFromProto(execDoneReq.Execution)

	d.logger.Debug().
		Any("execution", execution.Key()).
		Str("output", string(execution.Output)).
		Msg("fsm: Setting execution")
	_, err := d.store.SetExecutionDone(execution)

	return err
}

func (d *raftFSM) applySetExecution(buf []byte) interface{} {
	var pbex sxproto.Execution
	if err := proto.Unmarshal(buf, &pbex); err != nil {
		return err
	}
	execution := NewExecutionFromProto(&pbex)
	key, err := d.store.SetExecution(execution)
	if err != nil {
		return err
	}
	return key
}

// Snapshot returns a snapshot of the key-value store. We wrap
// the things we need in fsmSnapshot and then send that over to Persist.
// Persist encodes the needed data from fsmSnapshot and transport it to
// Restore where the necessary data is replicated into the finite state machine.
// This allows the consensus algorithm to truncate the replicated log.
func (d *raftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{store: d.store}, nil
}

// Restore stores the key-value store to a previous state.
func (d *raftFSM) Restore(r io.ReadCloser) error {
	defer r.Close()
	return d.store.Restore(r)
}

type fsmSnapshot struct {
	store Storage
}

func (d *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := d.store.Snapshot(sink); err != nil {
		_ = sink.Cancel()
		return err
	}

	// Close the sink.
	if err := sink.Close(); err != nil {
		return err
	}

	return nil
}

func (d *fsmSnapshot) Release() {}
