package job

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/sine-io/sinx/internal/common/ntime"
	"github.com/sine-io/sinx/internal/common/plugin"
)

// JobID is value object of the job.
type JobID string

// Job is the main entity of the job domain.
// It contains all the information needed to run a job.
type Job struct {
	// Job id. Must be unique, it's a copy of name.
	ID JobID `json:"id"`

	// Job name. Must be unique, acts as the id.
	Name string `json:"name"`

	// Display name of the job. If present, displayed instead of the name
	DisplayName string `json:"displayname"`

	// The timezone where the cron expression will be evaluated in.
	// Empty means local time.
	Timezone string `json:"timezone"`

	// Cron expression for the job. When to run the job.
	Schedule string `json:"schedule"`

	// Arbitrary string indicating the owner of the job.
	Owner string `json:"owner"`

	// Email address to use for notifications.
	OwnerEmail string `json:"owner_email"`

	// Number of successful executions of this job.
	SuccessCount int `json:"success_count"`

	// Number of errors running this job.
	ErrorCount int `json:"error_count"`

	// Last time this job executed successfully.
	LastSuccess ntime.NullableTime `json:"last_success"`

	// Last time this job failed.
	LastError ntime.NullableTime `json:"last_error"`

	// Is this job disabled?
	Disabled bool `json:"disabled"`

	// Tags of the target servers to run this job against.
	Tags map[string]string `json:"tags"`

	// Job metadata describes the job and allows filtering from the API.
	Metadata map[string]string `json:"metadata"`

	// Number of times to retry a job that failed an execution.
	Retries uint `json:"retries"`

	// Processors to use for this job.
	Processors map[string]plugin.Config `json:"processors"`

	// Executor plugin to be used in this job.
	Executor string `json:"executor"`

	// Configuration arguments for the specific executor.
	ExecutorConfig plugin.ExecutorPluginConfig `json:"executor_config"`

	// Computed job status.
	Status JobStatus `json:"status"`

	// Concurrency policy for this job (allow, forbid).
	Concurrency ConcurrencyStatus `json:"concurrency"`

	// Computed next execution.
	Next time.Time `json:"next"`

	// Delete the job after the first successful execution.
	Ephemeral bool `json:"ephemeral"`

	// The job will not be executed after this time.
	ExpiresAt ntime.NullableTime `json:"expires_at"`

	// Jobs that are dependent upon this one will be run after this job runs.
	DependentJobs []string `json:"dependent_jobs"` // want to remove in the future

	// Job pointer that are dependent upon this one
	ChildJobs []*Job `json:"-"` // want to remove in the future

	// Job id of job that this job is dependent upon.
	ParentJob string `json:"parent_job"` // want to remove in the future

	// logger
	logger zerolog.Logger `json:"-"`
}

// JobStatus is the value object of the job status.
type JobStatus string

const (
	// JobNotSet is the initial job status.
	JobNotSet JobStatus = ""
	// JobSuccess is status of a job whose last run was a success.
	JobSuccess JobStatus = "success"
	// JobRunning is status of a job whose last run has not finished.
	JobRunning JobStatus = "running"
	// JobFailed is status of a job whose last run was not successful on any nodes.
	JobFailed JobStatus = "failed"
	// JobPartiallyFailed is status of a job whose last run was successful on only some nodes.
	JobPartiallyFailed JobStatus = "partially_failed"
)

type ConcurrencyStatus string

const (
	// ConcurrencyAllow allows a job to execute concurrency.
	ConcurrencyAllow ConcurrencyStatus = "allow"
	// ConcurrencyForbid forbids a job from executing concurrency.
	ConcurrencyForbid ConcurrencyStatus = "forbid"
)
