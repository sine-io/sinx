package definition

import (
	"context"

	sxjob "github.com/sine-io/sinx/internal/job"
)

type Scheduler interface {
	Start(jobs []*sxjob.Job) error
	Stop() context.Context
	Restart(jobs []*sxjob.Job)
	AddJob(job *sxjob.Job) error
	RemoveJob(jobName string)
	GetCronEntryJob(jobName string) (sxjob.CronEntryJob, bool)
	ClearCron()
	Started() bool
}
