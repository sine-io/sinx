package definition

import (
	"io"

	sxexec "github.com/sine-io/sinx/internal/execution"
	sxjob "github.com/sine-io/sinx/internal/job"
)

var (
	JobDBInstance JobDB
)

// JobDB is the interface that should be used by any
// storage engine implemented for dkron. It contains the
// minimum set of operations that are needed to have a working
// dkron store.
type JobDB interface {
	SetJob(job *sxjob.Job, copyDependentJobs bool) error
	DeleteJob(name string) (*sxjob.Job, error)
	SetExecution(execution *sxexec.Execution) (string, error)
	SetExecutionDone(execution *sxexec.Execution) (bool, error)
	GetJobs(options *sxjob.JobOptions) ([]*sxjob.Job, error)
	GetJob(name string, options *sxjob.JobOptions) (*sxjob.Job, error)
	GetExecutions(jobName string, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error)
	GetExecutionGroup(execution *sxexec.Execution, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error)
	GetGroupedExecutions(jobName string, opts *sxexec.ExecutionOptions) (map[int64][]*sxexec.Execution, []int64, error)
	Shutdown() error
	Snapshot(w io.WriteCloser) error
	Restore(r io.ReadCloser) error
}
