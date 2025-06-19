package agent

import (
	"context"
)

type Scheduler interface {
	Start(jobs []*Job) error
	Stop() context.Context
	Restart(jobs []*Job)
	AddJob(job *Job) error
	RemoveJob(jobName string)
	GetCronEntryJob(jobName string) (CronEntryJob, bool)
	ClearCron()
	Started() bool
}
