package agent

import (
	"io"

	sxexec "github.com/sine-io/sinx/internal/execution"
)

// Storage is the interface that should be used by any
// storage engine implemented for sinx. It contains the
// minimum set of operations that are needed to have a working
// sinx store.
type Storage interface {
	SetJob(job *Job, copyDependentJobs bool) error
	DeleteJob(name string) (*Job, error)
	GetJobs(options *JobOptions) ([]*Job, error)
	GetJob(name string, options *JobOptions) (*Job, error)

	SetExecution(execution *sxexec.Execution) (string, error)
	SetExecutionDone(execution *sxexec.Execution) (bool, error)
	GetExecutions(jobName string, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error)
	GetExecutionGroup(execution *sxexec.Execution, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error)
	GetGroupedExecutions(jobName string, opts *sxexec.ExecutionOptions) (map[int64][]*sxexec.Execution, []int64, error)

	Shutdown() error
	Snapshot(w io.WriteCloser) error
	Restore(r io.ReadCloser) error
}
