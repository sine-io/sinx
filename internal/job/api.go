package job

import (
	"github.com/sine-io/sinx/internal/execution"
)

// JobExecutor 定义执行任务所需的接口方法
type JobExecutor interface {
	Execute(jobName string, execution *execution.Execution) (*execution.Execution, error)
}

// JobStorage 定义任务存储所需的接口方法
type JobStorage interface {
	GetJob(name string) (*Job, error)
	SetJob(job *Job) error
	DeleteJob(name string) error
	GetJobs() ([]*Job, error)
}
