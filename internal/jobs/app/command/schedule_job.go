package command

import (
	"github.com/sine-io/sinx/internal/common/decorator"
	"github.com/sine-io/sinx/internal/job/domain/job"
)

type ScheduleJob struct {
	JobID string
}

type ScheduleJobHandler decorator.CommandHandler[ScheduleJob]

type scheduleJobHandler struct {
	repo       job.Repository
	xxxService XxxService
	yyyServic  YyyServic
}
