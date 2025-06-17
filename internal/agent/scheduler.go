package agent

import (
	"context"
	"errors"
	"expvar"
	"strings"
	"sync"

	"github.com/armon/go-metrics"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"

	sxextcron "github.com/sine-io/sinx/extcron"
	sxjob "github.com/sine-io/sinx/internal/job"
)

var (
	cronInspect      = expvar.NewMap("cron_entries")
	schedulerStarted = expvar.NewInt("scheduler_started")

	// ErrScheduleParse is the error returned when the schedule parsing fails.
	ErrScheduleParse = errors.New("can't parse job schedule")
)

// CronScheduler represents a scheduler instance, it stores the cron engine
// and the related parameters.
type CronScheduler struct {
	// mu is to prevent concurrent edits to Cron and Started
	mu      sync.RWMutex
	Cron    *cron.Cron
	started bool
	logger  zerolog.Logger
}

// NewCronScheduler creates a new Scheduler instance
func NewCronScheduler() *CronScheduler {
	schedulerStarted.Set(0)
	return &CronScheduler{
		Cron:    cron.New(cron.WithParser(sxextcron.NewParser())),
		started: false,
		logger:  zerolog.New(zerolog.NewConsoleWriter()),
	}
}

func (s *CronScheduler) WithLogger(logger *zerolog.Logger) *CronScheduler {
	s.logger = logger.Hook()

	return s
}

// Start the cron scheduler, adding its corresponding jobs and
// executing them on time.
func (s *CronScheduler) Start(jobs []*sxjob.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("scheduler: cron already started, should be stopped first")
	}
	s.ClearCron()

	metrics.IncrCounter([]string{"scheduler", "start"}, 1)

	for _, job := range jobs {
		if err := s.AddJob(job); err != nil {
			return err
		}
	}
	s.Cron.Start()
	s.started = true
	schedulerStarted.Set(1)

	return nil
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
// A context is returned so the caller can wait for running jobs to complete.
func (s *CronScheduler) Stop() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := s.Cron.Stop()
	if s.started {
		s.logger.Debug().Msg("scheduler: Stopping scheduler")
		s.started = false

		// expvars
		cronInspect.Do(func(kv expvar.KeyValue) {
			kv.Value = nil
		})
	}
	schedulerStarted.Set(0)
	return ctx
}

// Restart the scheduler
func (s *CronScheduler) Restart(jobs []*sxjob.Job) {
	// Stop the scheduler, running jobs will continue to finish but we
	// can not actively wait for them blocking the execution here.
	s.Stop()

	if err := s.Start(jobs); err != nil {
		s.logger.Fatal().Err(err).Send()
	}
}

// ClearCron clears the cron scheduler
func (s *CronScheduler) ClearCron() {
	for _, e := range s.Cron.Entries() {
		if j, ok := e.Job.(*sxjob.Job); !ok {
			s.logger.Error().
				Msgf("scheduler: Failed to cast job to *Job found type %T and removing it", e.Job)
			s.Cron.Remove(e.ID)
		} else {
			s.RemoveJob(j.Name)
		}
	}
}

// Started will safely return if the scheduler is started or not
func (s *CronScheduler) Started() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.started
}

// GetCronEntryJob returns a CronEntryJob object from a snapshot in
// the current time, and whether or not the entry was found.
func (s *CronScheduler) GetCronEntryJob(jobName string) (sxjob.CronEntryJob, bool) {
	for _, e := range s.Cron.Entries() {
		if j, ok := e.Job.(*sxjob.Job); !ok {
			s.logger.Error().
				Msgf("scheduler: Failed to cast job to *Job found type %T", e.Job)
		} else {
			if j.Name == jobName {
				return sxjob.CronEntryJob{
					Entry: &e,
					Job:   j,
				}, true
			}
		}
	}
	return sxjob.CronEntryJob{}, false
}

// AddJob Adds a job to the cron scheduler
func (s *CronScheduler) AddJob(job *sxjob.Job) error {
	// Check if the job is already set and remove it if exists
	if _, ok := s.GetCronEntryJob(job.Name); ok {
		s.RemoveJob(job.Name)
	}

	if job.Disabled || job.ParentJob != "" {
		return nil
	}

	s.logger.Debug().Str("job", job.Name).
		Msg("scheduler: Adding job to cron")

	// If Timezone is set on the job, and not explicitly in its schedule,
	// AND its not a descriptor (that don't support timezones), add the
	// timezone to the schedule so robfig/cron knows about it.
	schedule := job.ScheduleHash()
	if job.Timezone != "" &&
		!strings.HasPrefix(schedule, "@") &&
		!strings.HasPrefix(schedule, "TZ=") &&
		!strings.HasPrefix(schedule, "CRON_TZ=") {
		schedule = "CRON_TZ=" + job.Timezone + " " + schedule
	}

	_, err := s.Cron.AddJob(schedule, job)
	if err != nil {
		return err
	}

	cronInspect.Set(job.Name, job)
	metrics.IncrCounterWithLabels(
		[]string{"scheduler", "job_add"},
		1,
		[]metrics.Label{{Name: "job", Value: job.Name}},
	)

	return nil
}

// RemoveJob removes a job from the cron scheduler if it exists.
func (s *CronScheduler) RemoveJob(jobName string) {

	s.logger.Debug().Str("job", jobName).
		Msg("scheduler: Removing job from cron")

	if ej, ok := s.GetCronEntryJob(jobName); ok {
		s.Cron.Remove(ej.Entry.ID)
		cronInspect.Delete(jobName)
		metrics.IncrCounterWithLabels(
			[]string{"scheduler", "job_delete"},
			1,
			[]metrics.Label{{Name: "job", Value: jobName}},
		)
	}
}
