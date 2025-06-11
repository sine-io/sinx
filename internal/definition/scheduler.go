package definition

import (
	"context"

	sxjob "github.com/sine-io/sinx/internal/job"
)

type Scheduler interface {
	Start(jobs []*sxjob.Job) error
	Stop() context.Context
	AddJob(job *sxjob.Job) error
	RemoveJob(jobName string)
	Restart(jobs []*sxjob.Job)
	ClearCron()
	Started() bool
	GetCronEntryJob(jobName string) (sxjob.CronEntryJob, bool)
}
