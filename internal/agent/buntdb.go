package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/tidwall/buntdb"
	"google.golang.org/protobuf/proto"

	sxexec "github.com/sine-io/sinx/internal/execution"
	sxjob "github.com/sine-io/sinx/internal/job"
	sxproto "github.com/sine-io/sinx/types"
)

const (
	// MaxExecutions to maintain in the storage
	MaxExecutions = 100

	jobsPrefix       = "jobs"
	executionsPrefix = "executions"
)

var (
	// ErrDependentJobs is returned when deleting a job that has dependent jobs
	ErrDependentJobs = errors.New("store: could not delete job with dependent jobs, delete childs first")
)

type kv struct {
	Key   string
	Value []byte
}

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

// BuntJobDB is the local implementation of the JobDB interface.
// It gives sinx the ability to manipulate its embedded storage
// BuntDB.
type BuntJobDB struct {
	db   *buntdb.DB
	lock *sync.Mutex

	logger zerolog.Logger
}

// NewBuntJobDB creates a new NewBuntJobDB instance.
func NewBuntJobDB() (*BuntJobDB, error) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return nil, err
	}
	_ = db.CreateIndex("name", jobsPrefix+":*", buntdb.IndexJSON("name"))
	_ = db.CreateIndex("started_at", executionsPrefix+":*", buntdb.IndexJSON("started_at"))
	_ = db.CreateIndex("finished_at", executionsPrefix+":*", buntdb.IndexJSON("finished_at"))
	_ = db.CreateIndex("attempt", executionsPrefix+":*", buntdb.IndexJSON("attempt"))
	_ = db.CreateIndex("displayname", jobsPrefix+":*", buntdb.IndexJSON("displayname"))
	_ = db.CreateIndex("schedule", jobsPrefix+":*", buntdb.IndexJSON("schedule"))
	_ = db.CreateIndex("success_count", jobsPrefix+":*", buntdb.IndexJSON("success_count"))
	_ = db.CreateIndex("error_count", jobsPrefix+":*", buntdb.IndexJSON("error_count"))
	_ = db.CreateIndex("last_success", jobsPrefix+":*", buntdb.IndexJSON("last_success"))
	_ = db.CreateIndex("last_error", jobsPrefix+":*", buntdb.IndexJSON("last_error"))
	_ = db.CreateIndex("next", jobsPrefix+":*", buntdb.IndexJSON("next"))

	bunt := &BuntJobDB{
		db:   db,
		lock: &sync.Mutex{},
		// set default logger, we should use WithLogger to set your own logger.
		logger: zerolog.New(zerolog.NewConsoleWriter()),
	}

	return bunt, nil
}

// WithLogger sets the logger for the BuntJobDB instance.
func (bjd *BuntJobDB) WithLogger(logger *zerolog.Logger) *BuntJobDB {
	bjd.logger = logger.Hook()

	return bjd
}

// SetJob stores a job in the storage
func (bjd *BuntJobDB) SetJob(job *sxjob.Job, copyDependentJobs bool) error {
	var pbej sxproto.Job
	var ej *sxjob.Job

	if err := job.Validate(); err != nil {
		return err
	}

	// Abort if parent not found before committing job to the store
	if job.ParentJob != "" {
		if j, _ := bjd.GetJob(job.ParentJob, nil); j == nil {
			return sxjob.ErrParentJobNotFound
		}
	}

	err := bjd.db.Update(func(tx *buntdb.Tx) error {
		// Get if the requested job already exist
		err := bjd.getJobTxFunc(job.Name, &pbej)(tx)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}

		ej = sxjob.NewJobFromProto(&pbej)

		if ej.Name != "" {
			// When the job runs, these status vars are updated
			// otherwise use the ones that are stored
			if ej.LastError.After(job.LastError) {
				job.LastError = ej.LastError
			}
			if ej.LastSuccess.After(job.LastSuccess) {
				job.LastSuccess = ej.LastSuccess
			}
			if ej.SuccessCount > job.SuccessCount {
				job.SuccessCount = ej.SuccessCount
			}
			if ej.ErrorCount > job.ErrorCount {
				job.ErrorCount = ej.ErrorCount
			}
			if len(ej.DependentJobs) != 0 && copyDependentJobs {
				job.DependentJobs = ej.DependentJobs
			}
			if ej.Status != "" {
				job.Status = ej.Status
			}
		}

		if job.Schedule != ej.Schedule {
			job.Next, err = job.GetNext()
			if err != nil {
				return err
			}
		} else {
			// If coming from a backup us the previous value, don't allow overwriting this
			if job.Next.Before(ej.Next) {
				job.Next = ej.Next
			}
		}

		pbj := job.ToProto()
		if err := bjd.setJobTxFunc(pbj)(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// If the parent job changed update the parents of the old (if any) and new jobs
	if job.ParentJob != ej.ParentJob {
		if err := bjd.removeFromParent(ej); err != nil {
			return err
		}
		if err := bjd.addToParent(job); err != nil {
			return err
		}
	}

	return nil
}

// SetExecutionDone saves the execution and updates the job with the corresponding
// results
func (bjd *BuntJobDB) SetExecutionDone(execution *sxexec.Execution) (bool, error) {
	err := bjd.db.Update(func(tx *buntdb.Tx) error {
		// Load the job from the store
		var pbj sxproto.Job
		if err := bjd.getJobTxFunc(execution.JobName, &pbj)(tx); err != nil {
			if err == buntdb.ErrNotFound {
				bjd.logger.Warn().Err(ErrExecutionDoneForDeletedJob).Send()
				return ErrExecutionDoneForDeletedJob
			}
			bjd.logger.Fatal().Err(err).Send()
			return err
		}

		key := fmt.Sprintf("%s:%s:%s", executionsPrefix, execution.JobName, execution.Key())

		// Save the execution to store
		pbe := execution.ToProto()
		if err := bjd.setExecutionTxFunc(key, pbe)(tx); err != nil {
			return err
		}

		if pbe.Success {
			pbj.LastSuccess.HasValue = true
			pbj.LastSuccess.Time = pbe.FinishedAt
			pbj.SuccessCount++
		} else {
			pbj.LastError.HasValue = true
			pbj.LastError.Time = pbe.FinishedAt
			pbj.ErrorCount++
		}

		status, err := bjd.computeStatus(pbj.Name, pbe.Group, tx)
		if err != nil {
			return err
		}
		pbj.Status = status

		if err := bjd.setJobTxFunc(&pbj)(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		bjd.logger.Error().Err(err).Msg("store: Error in SetExecutionDone")
		return false, err
	}

	return true, nil
}

// GetJobs returns all jobs
func (bjd *BuntJobDB) GetJobs(options *sxjob.JobOptions) ([]*sxjob.Job, error) {
	if options == nil {
		options = &sxjob.JobOptions{
			Sort: "name",
		}
	}

	jobs := make([]*sxjob.Job, 0)
	jobsFn := func(key, item string) bool {
		var pbj sxproto.Job
		// [TODO] This condition is temporary while we migrate to JSON marshalling for jobs
		// so we can use BuntDb indexes. To be removed in future versions.
		if err := proto.Unmarshal([]byte(item), &pbj); err != nil {
			if err := json.Unmarshal([]byte(item), &pbj); err != nil {
				return false
			}
		}
		job := sxjob.NewJobFromProto(&pbj)

		if options == nil ||
			(options.Metadata == nil || len(options.Metadata) == 0 || bjd.jobHasMetadata(job, options.Metadata)) &&
				(options.Query == "" || strings.Contains(job.Name, options.Query) || strings.Contains(job.DisplayName, options.Query)) &&
				(options.Disabled == "" || strconv.FormatBool(job.Disabled) == options.Disabled) &&
				((options.Status == "untriggered" && job.Status == "") || (options.Status == "" || job.Status == options.Status)) {

			jobs = append(jobs, job)
		}
		return true
	}

	err := bjd.db.View(func(tx *buntdb.Tx) error {
		var err error
		if options.Order == "DESC" {
			err = tx.Descend(options.Sort, jobsFn)
		} else {
			err = tx.Ascend(options.Sort, jobsFn)
		}
		return err
	})

	return jobs, err
}

// GetJob finds and return a Job from the store
func (bjd *BuntJobDB) GetJob(name string, options *sxjob.JobOptions) (*sxjob.Job, error) {
	var pbj sxproto.Job

	err := bjd.db.View(bjd.getJobTxFunc(name, &pbj))
	if err != nil {
		return nil, err
	}

	job := sxjob.NewJobFromProto(&pbj)

	return job, nil
}

// DeleteJob deletes the given job from the store, along with
// all its executions and references to it.
func (bjd *BuntJobDB) DeleteJob(name string) (*sxjob.Job, error) {
	var job *sxjob.Job
	err := bjd.db.Update(func(tx *buntdb.Tx) error {
		// Get the job
		var pbj sxproto.Job
		if err := bjd.getJobTxFunc(name, &pbj)(tx); err != nil {
			return err
		}
		// Check if the job has dependent jobs
		// and return an error indicating to remove childs
		// first.
		if len(pbj.DependentJobs) > 0 {
			return ErrDependentJobs
		}
		job = sxjob.NewJobFromProto(&pbj)

		if err := bjd.deleteExecutionsTxFunc(name)(tx); err != nil {
			return err
		}

		_, err := tx.Delete(fmt.Sprintf("%s:%s", jobsPrefix, name))
		return err
	})
	if err != nil {
		return nil, err
	}

	// If the transaction succeeded, remove from parent
	if job.ParentJob != "" {
		if err := bjd.removeFromParent(job); err != nil {
			return nil, err
		}
	}

	return job, nil
}

// GetExecutions returns the executions given a Job name.
func (bjd *BuntJobDB) GetExecutions(jobName string, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error) {
	prefix := fmt.Sprintf("%s:%s:", executionsPrefix, jobName)

	kvs, err := bjd.list(prefix, true, opts)
	if err != nil {
		return nil, err
	}

	return bjd.unmarshalExecutions(kvs, opts.Timezone)
}

// GetExecutionGroup returns all executions in the same group of a given execution
func (bjd *BuntJobDB) GetExecutionGroup(execution *sxexec.Execution, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error) {
	res, err := bjd.GetExecutions(execution.JobName, opts)
	if err != nil {
		return nil, err
	}

	var executions []*sxexec.Execution
	for _, ex := range res {
		if ex.Group == execution.Group {
			executions = append(executions, ex)
		}
	}
	return executions, nil
}

// GetGroupedExecutions returns executions for a job grouped and with an ordered index
// to facilitate access.
func (bjd *BuntJobDB) GetGroupedExecutions(jobName string, opts *sxexec.ExecutionOptions) (map[int64][]*sxexec.Execution, []int64, error) {
	execs, err := bjd.GetExecutions(jobName, opts)
	if err != nil {
		return nil, nil, err
	}
	groups := make(map[int64][]*sxexec.Execution)
	for _, exec := range execs {
		groups[exec.Group] = append(groups[exec.Group], exec)
	}

	// Build a separate data structure to show in order
	var byGroup int64arr
	for key := range groups {
		byGroup = append(byGroup, key)
	}
	sort.Sort(sort.Reverse(byGroup))

	return groups, byGroup, nil
}

// SetExecution Save a new execution and returns the key of the new saved item or an error.
func (bjd *BuntJobDB) SetExecution(execution *sxexec.Execution) (string, error) {
	pbe := execution.ToProto()
	key := fmt.Sprintf("%s:%s:%s", executionsPrefix, execution.JobName, execution.Key())

	bjd.logger.Debug().
		Str("job", execution.JobName).
		Str("execution", key).
		Str("finished", execution.FinishedAt.String()).
		Msg("store: Setting key")

	err := bjd.db.Update(bjd.setExecutionTxFunc(key, pbe))

	if err != nil {

		bjd.logger.Debug().
			Str("job", execution.JobName).
			Str("execution", key).
			Msg("store: Failed to set key")

		return "", err
	}

	execs, err := bjd.GetExecutions(execution.JobName, &sxexec.ExecutionOptions{})
	if err != nil && err != buntdb.ErrNotFound {
		bjd.logger.Error().
			Err(err).
			Str("job", execution.JobName).
			Msg("store: Error getting executions for job")
	}

	// Delete all execution results over the limit, starting from olders
	if len(execs) > MaxExecutions {
		//sort the array of all execution groups by StartedAt time
		sort.Slice(execs, func(i, j int) bool {
			return execs[i].StartedAt.Before(execs[j].StartedAt)
		})

		for i := 0; i < len(execs)-MaxExecutions; i++ {
			bjd.logger.Debug().
				Str("job", execs[i].JobName).
				Str("execution", execs[i].Key()).
				Msg("store: to delete key")

			err = bjd.db.Update(func(tx *buntdb.Tx) error {
				k := fmt.Sprintf("%s:%s:%s", executionsPrefix, execs[i].JobName, execs[i].Key())
				_, err := tx.Delete(k)
				return err
			})
			if err != nil {
				bjd.logger.Error().
					Err(err).
					Str("execution", execs[i].Key()).
					Msg("store: Error trying to delete overflowed execution")
			}
		}
	}

	return key, nil
}

// Shutdown close the KV store
func (bjd *BuntJobDB) Shutdown() error {
	return bjd.db.Close()
}

// Snapshot creates a backup of the data stored in BuntDB
func (bjd *BuntJobDB) Snapshot(w io.WriteCloser) error {
	return bjd.db.Save(w)
}

// Restore load data created with backup in to Bunt
func (bjd *BuntJobDB) Restore(r io.ReadCloser) error {
	return bjd.db.Load(r)
}

// DB is the getter for the BuntDB instance
// TODO: unused.
func (bjd *BuntJobDB) DB() *buntdb.DB {
	return bjd.db
}

func (bjd *BuntJobDB) setJobTxFunc(pbj *sxproto.Job) func(tx *buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		jobKey := fmt.Sprintf("%s:%s", jobsPrefix, pbj.Name)

		jb, err := json.Marshal(pbj)
		if err != nil {
			return err
		}
		bjd.logger.Debug().Str("job", pbj.Name).Msg("store: Setting job")

		if _, _, err := tx.Set(jobKey, string(jb), nil); err != nil {
			return err
		}

		return nil
	}
}

// This will allow reuse this code to avoid nesting transactions
func (bjd *BuntJobDB) getJobTxFunc(name string, pbj *sxproto.Job) func(tx *buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		item, err := tx.Get(fmt.Sprintf("%s:%s", jobsPrefix, name))
		if err != nil {
			return err
		}

		// [TODO] This condition is temporary while we migrate to JSON marshalling for jobs
		// so we can use BuntDb indexes. To be removed in future versions.
		if err := proto.Unmarshal([]byte(item), pbj); err != nil {
			if err := json.Unmarshal([]byte(item), pbj); err != nil {
				return err
			}
		}

		bjd.logger.Debug().Str("job", pbj.Name).Msg("store: Retrieved job from datastore")

		return nil
	}
}

// Removes the given job from its parent.
// Does nothing if nil is passed as child.
func (bjd *BuntJobDB) removeFromParent(child *sxjob.Job) error {
	// Do nothing if no job was given or job has no parent
	if child == nil || child.ParentJob == "" {
		return nil
	}

	parent, err := child.GetParent(bjd)
	if err != nil {
		return err
	}

	// Remove all occurrences from the parent, not just one.
	// Due to an old bug (in v1), a parent can have the same child more than once.
	djs := []string{}
	for _, djn := range parent.DependentJobs {
		if djn != child.Name {
			djs = append(djs, djn)
		}
	}
	parent.DependentJobs = djs
	if err := bjd.SetJob(parent, false); err != nil {
		return err
	}

	return nil
}

// Adds the given job to its parent.
func (bjd *BuntJobDB) addToParent(child *sxjob.Job) error {
	// Do nothing if job has no parent
	if child.ParentJob == "" {
		return nil
	}

	parent, err := child.GetParent(bjd)
	if err != nil {
		return err
	}

	parent.DependentJobs = append(parent.DependentJobs, child.Name)
	if err := bjd.SetJob(parent, false); err != nil {
		return err
	}

	return nil
}

func (bjd *BuntJobDB) setExecutionTxFunc(key string, pbe *sxproto.Execution) func(tx *buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		// Get previous execution
		i, err := tx.Get(key)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		// Do nothing if a previous execution exists and is
		// more recent, avoiding non ordered execution set
		if i != "" {
			var p sxproto.Execution
			// TODO: This condition is temporary while we migrate to JSON marshalling for executions
			// so we can use BuntDb indexes. To be removed in future versions.
			if err := proto.Unmarshal([]byte(i), &p); err != nil {
				if err := json.Unmarshal([]byte(i), &p); err != nil {
					return err
				}
			}
			// Compare existing execution
			if p.GetFinishedAt().Seconds > pbe.GetFinishedAt().Seconds {
				return nil
			}
		}

		eb, err := json.Marshal(pbe)
		if err != nil {
			return err
		}

		_, _, err = tx.Set(key, string(eb), nil)
		return err
	}
}

// deleteExecutionsTxFunc removes all executions of a job
func (bjd *BuntJobDB) deleteExecutionsTxFunc(jobName string) func(tx *buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		var delkeys []string
		prefix := fmt.Sprintf("%s:%s", executionsPrefix, jobName)
		if err := tx.Ascend("", func(key, value string) bool {
			if strings.HasPrefix(key, prefix) {
				delkeys = append(delkeys, key)
			}
			return true
		}); err != nil {
			return err
		}

		for _, k := range delkeys {
			_, _ = tx.Delete(k)
		}

		return nil
	}
}

func (bjd *BuntJobDB) list(prefix string, checkRoot bool, opts *sxexec.ExecutionOptions) ([]kv, error) {
	var found bool
	kvs := []kv{}

	err := bjd.db.View(bjd.listTxFunc(prefix, &kvs, &found, opts))
	if err == nil && !found && checkRoot {
		return nil, buntdb.ErrNotFound
	}

	return kvs, err
}

func (bjd *BuntJobDB) listTxFunc(prefix string, kvs *[]kv, found *bool, opts *sxexec.ExecutionOptions) func(tx *buntdb.Tx) error {
	fnc := func(key, value string) bool {
		if strings.HasPrefix(key, prefix) {
			*found = true
			// ignore self in listing
			if !bytes.Equal(trimDirectoryKey([]byte(key)), []byte(prefix)) {
				kv := kv{Key: key, Value: []byte(value)}
				*kvs = append(*kvs, kv)
			}
		}
		return true
	}

	return func(tx *buntdb.Tx) (err error) {
		if opts.Order == "DESC" {
			err = tx.Descend(opts.Sort, fnc)
		} else {
			err = tx.Ascend(opts.Sort, fnc)
		}
		return err
	}
}

func (bjd *BuntJobDB) unmarshalExecutions(items []kv, timezone *time.Location) ([]*sxexec.Execution, error) {
	var executions []*sxexec.Execution
	for _, item := range items {
		var pbe sxproto.Execution

		// [TODO] This condition is temporary while we migrate to JSON marshalling for jobs
		// so we can use BuntDb indexes. To be removed in future versions.
		if err := proto.Unmarshal([]byte(item.Value), &pbe); err != nil {
			if err := json.Unmarshal(item.Value, &pbe); err != nil {
				bjd.logger.Debug().Err(err).Str("key", item.Key).
					Msg("store: error unmarshaling JSON")

				return nil, err
			}
		}
		execution := sxexec.NewExecutionFromProto(&pbe)
		if timezone != nil {
			execution.FinishedAt = execution.FinishedAt.In(timezone)
			execution.StartedAt = execution.StartedAt.In(timezone)
		}
		executions = append(executions, execution)
	}
	return executions, nil
}

func (bjd *BuntJobDB) computeStatus(jobName string, exGroup int64, tx *buntdb.Tx) (string, error) {
	// compute job status based on execution group
	kvs := []kv{}
	found := false
	prefix := fmt.Sprintf("%s:%s:", executionsPrefix, jobName)

	if err := bjd.listTxFunc(prefix, &kvs, &found, &sxexec.ExecutionOptions{})(tx); err != nil {
		return "", err
	}

	execs, err := bjd.unmarshalExecutions(kvs, nil)
	if err != nil {
		return "", err
	}

	var executions []*sxexec.Execution
	for _, ex := range execs {
		if ex.Group == exGroup {
			executions = append(executions, ex)
		}
	}

	success := 0
	failed := 0

	var status string
	for _, ex := range executions {
		if ex.Success {
			success = success + 1
		} else {
			failed = failed + 1
		}
	}

	if failed == 0 {
		status = sxjob.JobStatusSuccess
	} else if failed > 0 && success == 0 {
		status = sxjob.JobStatusFailed
	} else if failed > 0 && success > 0 {
		status = sxjob.JobStatusPartiallyFailed
	}

	return status, nil
}

func (bjd *BuntJobDB) jobHasMetadata(job *sxjob.Job, metadata map[string]string) bool {
	if job == nil || job.Metadata == nil || len(job.Metadata) == 0 {
		return false
	}

	for k, v := range metadata {
		if val, ok := job.Metadata[k]; !ok || v != val {
			return false
		}
	}

	return true
}

func trimDirectoryKey(key []byte) []byte {
	if isDirectoryKey(key) {
		return key[:len(key)-1]
	}

	return key
}

func isDirectoryKey(key []byte) bool {
	return len(key) > 0 && key[len(key)-1] == ':'
}
