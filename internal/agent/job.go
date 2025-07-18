package agent

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/tidwall/buntdb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sxextcron "github.com/sine-io/sinx/extcron"
	sxntime "github.com/sine-io/sinx/ntime"
	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

const (
	// JobStatusNotSet is the initial job status.
	JobStatusNotSet = ""
	// JobStatusSuccess is status of a job whose last run was a success.
	JobStatusSuccess = "success"
	// JobStatusRunning is status of a job whose last run has not finished.
	JobStatusRunning = "running"
	// JobStatusFailed is status of a job whose last run was not successful on any nodes.
	JobStatusFailed = "failed"
	// JobStatusPartiallyFailed is status of a job whose last run was successful on only some nodes.
	JobStatusPartiallyFailed = "partially_failed"

	// ConcurrencyAllow allows a job to execute concurrency.
	ConcurrencyAllow = "allow"
	// ConcurrencyForbid forbids a job from executing concurrency.
	ConcurrencyForbid = "forbid"

	// HashSymbol is the "magic" character used in scheduled to be replaced with a value based on job name
	HashSymbol = "~"
)

var (
	// ErrParentJobNotFound is returned when the parent job is not found.
	ErrParentJobNotFound = errors.New("specified parent job not found")
	// ErrNoAgent is returned when the job's agent is nil.
	ErrNoAgent = errors.New("no agent defined")
	// ErrSameParent is returned when the job's parent is itself.
	ErrSameParent = errors.New("the job can not have itself as parent")
	// ErrNoParent is returned when the job has no parent.
	ErrNoParent = errors.New("the job doesn't have a parent job set")
	// ErrNoCommand is returned when attempting to store a job that has no command.
	ErrNoCommand = errors.New("unspecified command for job")
	// ErrWrongConcurrency is returned when Concurrency is set to a non existing setting.
	ErrWrongConcurrency = errors.New("invalid concurrency policy value, use \"allow\" or \"forbid\"")
)

// Job describes a scheduled Job.
type Job struct {
	// Job id. Must be unique, it's a copy of name.
	ID string `json:"id"`

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
	LastSuccess sxntime.NullableTime `json:"last_success"`

	// Last time this job failed.
	LastError sxntime.NullableTime `json:"last_error"`

	// Is this job disabled?
	Disabled bool `json:"disabled"`

	// Tags of the target servers to run this job against.
	Tags map[string]string `json:"tags"`

	// Job metadata describes the job and allows filtering from the API.
	Metadata map[string]string `json:"metadata"`

	// Pointer to the calling agent.
	Agent *Agent `json:"-"`

	// Number of times to retry a job that failed an execution.
	Retries uint `json:"retries"`

	// Jobs that are dependent upon this one will be run after this job runs.
	DependentJobs []string `json:"dependent_jobs"`

	// Job pointer that are dependent upon this one
	ChildJobs []*Job `json:"-"`

	// Job id of job that this job is dependent upon.
	ParentJob string `json:"parent_job"`

	// Processors to use for this job.
	Processors map[string]sxplugin.Config `json:"processors"`

	// Concurrency policy for this job (allow, forbid).
	Concurrency string `json:"concurrency"`

	// Executor plugin to be used in this job.
	Executor string `json:"executor"`

	// Configuration arguments for the specific executor.
	ExecutorConfig sxplugin.ExecutorPluginConfig `json:"executor_config"`

	// Computed job status.
	Status string `json:"status"`

	// Computed next execution.
	Next time.Time `json:"next"`

	// Delete the job after the first successful execution.
	Ephemeral bool `json:"ephemeral"`

	// The job will not be executed after this time.
	ExpiresAt sxntime.NullableTime `json:"expires_at"`

	logger zerolog.Logger `json:"-"`
}

// NewJobFromProto create a new Job from a PB Job struct
func NewJobFromProto(in *sxproto.Job) *Job {
	job := &Job{
		ID:             in.Name,
		Name:           in.Name,
		DisplayName:    in.Displayname,
		Timezone:       in.Timezone,
		Schedule:       in.Schedule,
		Owner:          in.Owner,
		OwnerEmail:     in.OwnerEmail,
		SuccessCount:   int(in.SuccessCount),
		ErrorCount:     int(in.ErrorCount),
		Disabled:       in.Disabled,
		Tags:           in.Tags,
		Retries:        uint(in.Retries),
		DependentJobs:  in.DependentJobs,
		ParentJob:      in.ParentJob,
		Concurrency:    in.Concurrency,
		Executor:       in.Executor,
		ExecutorConfig: in.ExecutorConfig,
		Status:         in.Status,
		Metadata:       in.Metadata,
		Next:           in.GetNext().AsTime(),
		Ephemeral:      in.Ephemeral,
	}
	if in.GetLastSuccess().GetHasValue() {
		t := in.GetLastSuccess().GetTime().AsTime()
		job.LastSuccess.Set(t)
	}
	if in.GetLastError().GetHasValue() {
		t := in.GetLastError().GetTime().AsTime()
		job.LastError.Set(t)
	}
	if in.GetExpiresAt().GetHasValue() {
		t := in.GetExpiresAt().GetTime().AsTime()
		job.ExpiresAt.Set(t)
	}

	procs := make(map[string]sxplugin.Config)
	for k, v := range in.Processors {
		if len(v.Config) == 0 {
			v.Config = make(map[string]string)
		}
		procs[k] = v.Config
	}
	job.Processors = procs

	return job
}

// ToProto return the corresponding representation of this Job in proto struct
func (j *Job) ToProto() *sxproto.Job {
	lastSuccess := &sxproto.Job_NullableTime{
		HasValue: j.LastSuccess.HasValue(),
	}
	if j.LastSuccess.HasValue() {
		lastSuccess.Time = timestamppb.New(j.LastSuccess.Get())
	}
	lastError := &sxproto.Job_NullableTime{
		HasValue: j.LastError.HasValue(),
	}
	if j.LastError.HasValue() {
		lastError.Time = timestamppb.New(j.LastError.Get())
	}

	next := timestamppb.New(j.Next)

	expiresAt := &sxproto.Job_NullableTime{
		HasValue: j.ExpiresAt.HasValue(),
	}
	if j.ExpiresAt.HasValue() {
		expiresAt.Time = timestamppb.New(j.ExpiresAt.Get())
	}

	processors := make(map[string]*sxproto.PluginConfig)
	for k, v := range j.Processors {
		processors[k] = &sxproto.PluginConfig{Config: v}
	}
	return &sxproto.Job{
		Name:           j.Name,
		Displayname:    j.DisplayName,
		Timezone:       j.Timezone,
		Schedule:       j.Schedule,
		Owner:          j.Owner,
		OwnerEmail:     j.OwnerEmail,
		SuccessCount:   int32(j.SuccessCount),
		ErrorCount:     int32(j.ErrorCount),
		Disabled:       j.Disabled,
		Tags:           j.Tags,
		Retries:        uint32(j.Retries),
		DependentJobs:  j.DependentJobs,
		ParentJob:      j.ParentJob,
		Concurrency:    j.Concurrency,
		Processors:     processors,
		Executor:       j.Executor,
		ExecutorConfig: j.ExecutorConfig,
		Status:         j.Status,
		Metadata:       j.Metadata,
		LastSuccess:    lastSuccess,
		LastError:      lastError,
		Next:           next,
		Ephemeral:      j.Ephemeral,
		ExpiresAt:      expiresAt,
	}
}

// Friendly format a job
func (j *Job) String() string {
	return fmt.Sprintf("\"Job: %s, scheduled at: %s, tags:%v\"", j.Name, j.Schedule, j.Tags)
}

// GetParent returns the parent job of a job
func (j *Job) GetParent(store *BuntdbStore) (*Job, error) {
	if j.Name == j.ParentJob {
		return nil, ErrSameParent
	}

	if j.ParentJob == "" {
		return nil, ErrNoParent
	}

	parentJob, err := store.GetJob(j.ParentJob, nil)
	if err != nil {
		if err == buntdb.ErrNotFound {
			return nil, ErrParentJobNotFound
		}
		return nil, err

	}

	return parentJob, nil
}

// GetTimeLocation returns the time.Location based on the job's Timezone, or
// the default (UTC) if none is configured, or
// nil if an error occurred while creating the timezone from the property
func (j *Job) GetTimeLocation() *time.Location {
	loc, _ := time.LoadLocation(j.Timezone)
	return loc
}

// nameHash returns hash code of the job name
func (j *Job) nameHash() int {
	hash := 0
	for _, c := range j.Name {
		hash += int(c)
	}
	return hash
}

// ScheduleHash replaces H in the cron spec by a value derived from job Name
// such as "0 0 ~ * * *"
func (j *Job) ScheduleHash() string {
	spec := j.Schedule

	if !strings.Contains(spec, HashSymbol) {
		return spec
	}

	hash := j.nameHash()
	parts := strings.Split(spec, " ")
	partIndex := 0
	for index, part := range parts {
		if strings.HasPrefix(part, "@") {
			// this is a pre-defined scheduled, ignore everything
			return spec
		}
		if strings.HasPrefix(part, "TZ=") || strings.HasPrefix(part, "CRON_TZ=") {
			// do not increase partIndex
			continue
		}

		if strings.Contains(part, HashSymbol) {
			// mods taken in accordance with https://dkron.io/docs/usage/cron-spec/#cron-expression-format
			partHash := hash
			switch partIndex {
			case 2:
				partHash %= 24
			case 3:
				partHash = (partHash % 28) + 1
			case 4:
				partHash = (partHash % 12) + 1
			case 5:
				partHash %= 7
			default:
				partHash %= 60
			}
			parts[index] = strings.ReplaceAll(part, HashSymbol, strconv.Itoa(partHash))
		}

		partIndex++
	}

	return strings.Join(parts, " ")
}

// GetNext returns the job's next schedule from now
func (j *Job) GetNext() (time.Time, error) {
	if j.Schedule != "" {
		s, err := sxextcron.Parse(j.ScheduleHash())
		if err != nil {
			return time.Time{}, err
		}
		return s.Next(time.Now()), nil
	}

	return time.Time{}, nil
}

func (j *Job) isRunnable() bool {
	if j.Disabled || (j.ExpiresAt.HasValue() && time.Now().After(j.ExpiresAt.Get())) {
		j.logger.Debug().
			Str("job", j.Name).
			Msg("job: Skipping execution because job is disabled or expired")

		return false
	}

	if j.Agent.GlobalLock {
		j.logger.Warn().
			Str("job", j.Name).
			Msg("job: Skipping execution because active global lock")

		return false
	}

	if j.Concurrency == ConcurrencyForbid {
		exs, err := j.Agent.GetActiveExecutions()
		if err != nil {
			j.logger.Error().
				Err(err).
				Msg("job: Error querying for running executions")

			return false
		}

		for _, e := range exs {
			if e.JobName == j.Name {
				j.logger.Info().
					Str("job", j.Name).
					Str("concurrency", j.Concurrency).
					Str("job_status", j.Status).
					Msg("job: Skipping concurrent execution")

				return false
			}
		}
	}

	return true
}

// Validate validates whether all values in the job are acceptable.
func (j *Job) Validate() error {
	if j.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if valid, chr := isSlug(j.Name); !valid {
		return fmt.Errorf("name contains illegal character '%s'", chr)
	}

	if j.ParentJob == j.Name {
		return ErrSameParent
	}

	// Validate schedule, allow empty schedule if parent job set.
	if j.Schedule != "" || j.ParentJob == "" {
		if _, err := sxextcron.Parse(j.ScheduleHash()); err != nil {
			return fmt.Errorf("%s: %s", ErrScheduleParse.Error(), err)
		}
	}

	if j.Concurrency != ConcurrencyAllow && j.Concurrency != ConcurrencyForbid && j.Concurrency != "" {
		return ErrWrongConcurrency
	}

	// An empty string is a valid timezone for LoadLocation
	if _, err := time.LoadLocation(j.Timezone); err != nil {
		return err
	}

	if j.Executor == "shell" && j.ExecutorConfig["timeout"] != "" {
		_, err := time.ParseDuration(j.ExecutorConfig["timeout"])
		if err != nil {
			return fmt.Errorf("Error parsing job timeout value")
		}
	}

	return nil
}

// isSlug determines whether the given string is a proper value to be used as
// key in the backend store (a "slug"). If false, the 2nd return value
// will contain the first illegal character found.
func isSlug(candidate string) (bool, string) {
	// Allow only lower case letters (unicode), digits, underscore and dash.
	illegalCharPattern, _ := regexp.Compile(`[^\p{Ll}0-9_-]`)
	whyNot := illegalCharPattern.FindString(candidate)
	return whyNot == "", whyNot
}

// GenerateJobTree generate Job Tree
func GenerateJobTree(jobs []*Job) ([]*Job, error) {
	length := len(jobs)
	j := 0
	for i := 0; i < length; i++ {
		rejobs, isTopParentNodeFlag, err := findParentJobAndValidateJob(jobs, j)
		if err != nil {
			return nil, err
		}
		if isTopParentNodeFlag {
			j++
		}
		jobs = rejobs
	}
	return jobs, nil
}

func RecursiveSetJob(jobs []*Job) []string {
	result := make([]string, 0)
	for _, job := range jobs {
		err := job.Agent.GRPCClient.SetJob(job)
		if err != nil {
			result = append(result, "fail create "+job.Name)
			continue
		} else {
			result = append(result, "success create "+job.Name)
			if len(job.ChildJobs) > 0 {
				recursiveResult := RecursiveSetJob(job.ChildJobs)
				result = append(result, recursiveResult...)
			}
		}
	}
	return result
}

// findParentJobAndValidateJob...
func findParentJobAndValidateJob(jobs []*Job, index int) ([]*Job, bool, error) {
	childJob := jobs[index]
	// Validate job
	if err := childJob.Validate(); err != nil {
		return nil, false, err
	}
	if childJob.ParentJob == "" {
		return jobs, true, nil
	}
	for _, parentJob := range jobs {
		if parentJob.Name == childJob.Name {
			continue
		}
		if childJob.ParentJob == parentJob.Name {
			parentJob.ChildJobs = append(parentJob.ChildJobs, childJob)
			jobs = append(jobs[:index], jobs[index+1:]...)
			return jobs, false, nil
		}
		if len(parentJob.ChildJobs) > 0 {
			flag := findParentJobInChildJobs(parentJob.ChildJobs, childJob)
			if flag {
				jobs = append(jobs[:index], jobs[index+1:]...)
				return jobs, false, nil
			}
		}
	}
	return nil, false, ErrNoParent
}

func findParentJobInChildJobs(jobs []*Job, job *Job) bool {
	for _, parentJob := range jobs {
		if job.ParentJob == parentJob.Name {
			parentJob.ChildJobs = append(parentJob.ChildJobs, job)
			return true
		} else {
			if len(parentJob.ChildJobs) > 0 {
				flag := findParentJobInChildJobs(parentJob.ChildJobs, job)
				if flag {
					return true
				}
			}

		}
	}
	return false
}
