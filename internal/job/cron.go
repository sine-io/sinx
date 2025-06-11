package job

import "github.com/robfig/cron/v3"

type CronEntryJob struct {
	Entry *cron.Entry
	Job   *Job
}
