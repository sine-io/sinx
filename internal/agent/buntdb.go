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

// BuntdbStore is the local implementation of the Storage interface.
// It gives sinx the ability to manipulate its embedded storage
// BuntDB.
type BuntdbStore struct {
	db   *buntdb.DB
	lock *sync.Mutex

	logger zerolog.Logger
}

// NewBuntdbStore creates a new buntdb store instance.
func NewBuntdbStore() (*BuntdbStore, error) {
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

	bunt := &BuntdbStore{
		db:   db,
		lock: &sync.Mutex{},
		// set default logger, we should use WithLogger to set your own logger.
		logger: zerolog.New(zerolog.NewConsoleWriter()),
	}

	return bunt, nil
}

// WithLogger sets the logger for the BuntdbStore instance.
func (bs *BuntdbStore) WithLogger(logger *zerolog.Logger) *BuntdbStore {
	bs.logger = logger.Hook()

	return bs
}

// SetJob stores a job in the storage
func (bs *BuntdbStore) SetJob(job *Job, copyDependentJobs bool) error {
	var pbej sxproto.Job
	var ej *Job

	if err := job.Validate(); err != nil {
		return err
	}

	// Abort if parent not found before committing job to the store
	if job.ParentJob != "" {
		if j, _ := bs.GetJob(job.ParentJob, nil); j == nil {
			return ErrParentJobNotFound
		}
	}

	err := bs.db.Update(func(tx *buntdb.Tx) error {
		// Get if the requested job already exist
		err := bs.getJobTxFunc(job.Name, &pbej)(tx)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}

		ej = NewJobFromProto(&pbej)

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
		if err := bs.setJobTxFunc(pbj)(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// If the parent job changed update the parents of the old (if any) and new jobs
	if job.ParentJob != ej.ParentJob {
		if err := bs.removeFromParent(ej); err != nil {
			return err
		}
		if err := bs.addToParent(job); err != nil {
			return err
		}
	}

	return nil
}

// SetExecutionDone saves the execution and updates the job with the corresponding
// results
func (bs *BuntdbStore) SetExecutionDone(execution *sxexec.Execution) (bool, error) {
	err := bs.db.Update(func(tx *buntdb.Tx) error {
		// Load the job from the store
		var pbj sxproto.Job
		if err := bs.getJobTxFunc(execution.JobName, &pbj)(tx); err != nil {
			if err == buntdb.ErrNotFound {
				bs.logger.Warn().Err(ErrExecutionDoneForDeletedJob).Send()
				return ErrExecutionDoneForDeletedJob
			}
			bs.logger.Fatal().Err(err).Send()
			return err
		}

		key := fmt.Sprintf("%s:%s:%s", executionsPrefix, execution.JobName, execution.Key())

		// Save the execution to store
		pbe := execution.ToProto()
		if err := bs.setExecutionTxFunc(key, pbe)(tx); err != nil {
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

		status, err := bs.computeStatus(pbj.Name, pbe.Group, tx)
		if err != nil {
			return err
		}
		pbj.Status = status

		if err := bs.setJobTxFunc(&pbj)(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		bs.logger.Error().Err(err).Msg("store: Error in SetExecutionDone")
		return false, err
	}

	return true, nil
}

// GetJobs returns all jobs
func (bs *BuntdbStore) GetJobs(options *JobOptions) ([]*Job, error) {
	if options == nil {
		options = &JobOptions{
			Sort: "name",
		}
	}

	jobs := make([]*Job, 0)
	jobsFn := func(key, item string) bool {
		var pbj sxproto.Job
		// [TODO] This condition is temporary while we migrate to JSON marshalling for jobs
		// so we can use BuntDb indexes. To be removed in future versions.
		if err := proto.Unmarshal([]byte(item), &pbj); err != nil {
			if err := json.Unmarshal([]byte(item), &pbj); err != nil {
				return false
			}
		}
		job := NewJobFromProto(&pbj)

		if options == nil ||
			(len(options.Metadata) == 0 || bs.jobHasMetadata(job, options.Metadata)) &&
				(options.Query == "" || strings.Contains(job.Name, options.Query) || strings.Contains(job.DisplayName, options.Query)) &&
				(options.Disabled == "" || strconv.FormatBool(job.Disabled) == options.Disabled) &&
				((options.Status == "untriggered" && job.Status == "") || (options.Status == "" || job.Status == options.Status)) {

			jobs = append(jobs, job)
		}

		return true
	}

	err := bs.db.View(func(tx *buntdb.Tx) error {
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
func (bs *BuntdbStore) GetJob(name string, options *JobOptions) (*Job, error) {
	var pbj sxproto.Job

	err := bs.db.View(bs.getJobTxFunc(name, &pbj))
	if err != nil {
		return nil, err
	}

	job := NewJobFromProto(&pbj)

	return job, nil
}

// DeleteJob deletes the given job from the store, along with
// all its executions and references to it.
func (bs *BuntdbStore) DeleteJob(name string) (*Job, error) {
	var job *Job
	err := bs.db.Update(func(tx *buntdb.Tx) error {
		// Get the job
		var pbj sxproto.Job
		if err := bs.getJobTxFunc(name, &pbj)(tx); err != nil {
			return err
		}
		// Check if the job has dependent jobs
		// and return an error indicating to remove childs
		// first.
		if len(pbj.DependentJobs) > 0 {
			return ErrDependentJobs
		}
		job = NewJobFromProto(&pbj)

		if err := bs.deleteExecutionsTxFunc(name)(tx); err != nil {
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
		if err := bs.removeFromParent(job); err != nil {
			return nil, err
		}
	}

	return job, nil
}

// GetExecutions returns the executions given a Job name.
func (bs *BuntdbStore) GetExecutions(jobName string, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error) {
	prefix := fmt.Sprintf("%s:%s:", executionsPrefix, jobName)

	kvs, err := bs.list(prefix, true, opts)
	if err != nil {
		return nil, err
	}

	return bs.unmarshalExecutions(kvs, opts.Timezone)
}

// GetExecutionGroup returns all executions in the same group of a given execution
func (bs *BuntdbStore) GetExecutionGroup(execution *sxexec.Execution, opts *sxexec.ExecutionOptions) ([]*sxexec.Execution, error) {
	res, err := bs.GetExecutions(execution.JobName, opts)
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
func (bs *BuntdbStore) GetGroupedExecutions(jobName string, opts *sxexec.ExecutionOptions) (map[int64][]*sxexec.Execution, []int64, error) {
	execs, err := bs.GetExecutions(jobName, opts)
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
func (bs *BuntdbStore) SetExecution(execution *sxexec.Execution) (string, error) {
	pbe := execution.ToProto()
	key := fmt.Sprintf("%s:%s:%s", executionsPrefix, execution.JobName, execution.Key())

	bs.logger.Debug().
		Str("job", execution.JobName).
		Str("execution", key).
		Str("finished", execution.FinishedAt.String()).
		Msg("store: Setting key")

	err := bs.db.Update(bs.setExecutionTxFunc(key, pbe))

	if err != nil {

		bs.logger.Debug().
			Str("job", execution.JobName).
			Str("execution", key).
			Msg("store: Failed to set key")

		return "", err
	}

	execs, err := bs.GetExecutions(execution.JobName, &sxexec.ExecutionOptions{})
	if err != nil && err != buntdb.ErrNotFound {
		bs.logger.Error().
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
			bs.logger.Debug().
				Str("job", execs[i].JobName).
				Str("execution", execs[i].Key()).
				Msg("store: to delete key")

			err = bs.db.Update(func(tx *buntdb.Tx) error {
				k := fmt.Sprintf("%s:%s:%s", executionsPrefix, execs[i].JobName, execs[i].Key())
				_, err := tx.Delete(k)
				return err
			})
			if err != nil {
				bs.logger.Error().
					Err(err).
					Str("execution", execs[i].Key()).
					Msg("store: Error trying to delete overflowed execution")
			}
		}
	}

	return key, nil
}

// Shutdown close the KV store
func (bs *BuntdbStore) Shutdown() error {
	return bs.db.Close()
}

// Snapshot creates a backup of the data stored in BuntDB
func (bs *BuntdbStore) Snapshot(w io.WriteCloser) error {
	return bs.db.Save(w)
}

// Restore load data created with backup in to Bunt
func (bs *BuntdbStore) Restore(r io.ReadCloser) error {
	return bs.db.Load(r)
}

// DB is the getter for the BuntDB instance
// TODO: unused.
func (bs *BuntdbStore) DB() *buntdb.DB {
	return bs.db
}

func (bs *BuntdbStore) setJobTxFunc(pbj *sxproto.Job) func(tx *buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		jobKey := fmt.Sprintf("%s:%s", jobsPrefix, pbj.Name)

		jb, err := json.Marshal(pbj)
		if err != nil {
			return err
		}
		bs.logger.Debug().Str("job", pbj.Name).Msg("store: Setting job")

		if _, _, err := tx.Set(jobKey, string(jb), nil); err != nil {
			return err
		}

		return nil
	}
}

// This will allow reuse this code to avoid nesting transactions
func (bs *BuntdbStore) getJobTxFunc(name string, pbj *sxproto.Job) func(tx *buntdb.Tx) error {
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

		bs.logger.Debug().Str("job", pbj.Name).Msg("store: Retrieved job from datastore")

		return nil
	}
}

// Removes the given job from its parent.
// Does nothing if nil is passed as child.
func (bs *BuntdbStore) removeFromParent(child *Job) error {
	// Do nothing if no job was given or job has no parent
	if child == nil || child.ParentJob == "" {
		return nil
	}

	parent, err := child.GetParent(bs)
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
	if err := bs.SetJob(parent, false); err != nil {
		return err
	}

	return nil
}

// Adds the given job to its parent.
func (bs *BuntdbStore) addToParent(child *Job) error {
	// Do nothing if job has no parent
	if child.ParentJob == "" {
		return nil
	}

	parent, err := child.GetParent(bs)
	if err != nil {
		return err
	}

	parent.DependentJobs = append(parent.DependentJobs, child.Name)
	if err := bs.SetJob(parent, false); err != nil {
		return err
	}

	return nil
}

func (bs *BuntdbStore) setExecutionTxFunc(key string, pbe *sxproto.Execution) func(tx *buntdb.Tx) error {
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
func (bs *BuntdbStore) deleteExecutionsTxFunc(jobName string) func(tx *buntdb.Tx) error {
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

func (bs *BuntdbStore) list(prefix string, checkRoot bool, opts *sxexec.ExecutionOptions) ([]kv, error) {
	var found bool
	kvs := []kv{}

	err := bs.db.View(bs.listTxFunc(prefix, &kvs, &found, opts))
	if err == nil && !found && checkRoot {
		return nil, buntdb.ErrNotFound
	}

	return kvs, err
}

func (bs *BuntdbStore) listTxFunc(prefix string, kvs *[]kv, found *bool, opts *sxexec.ExecutionOptions) func(tx *buntdb.Tx) error {
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

func (bs *BuntdbStore) unmarshalExecutions(items []kv, timezone *time.Location) ([]*sxexec.Execution, error) {
	var executions []*sxexec.Execution
	for _, item := range items {
		var pbe sxproto.Execution

		// [TODO] This condition is temporary while we migrate to JSON marshalling for jobs
		// so we can use BuntDb indexes. To be removed in future versions.
		if err := proto.Unmarshal([]byte(item.Value), &pbe); err != nil {
			if err := json.Unmarshal(item.Value, &pbe); err != nil {
				bs.logger.Debug().Err(err).Str("key", item.Key).
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

func (bs *BuntdbStore) computeStatus(jobName string, exGroup int64, tx *buntdb.Tx) (string, error) {
	// compute job status based on execution group
	kvs := []kv{}
	found := false
	prefix := fmt.Sprintf("%s:%s:", executionsPrefix, jobName)

	if err := bs.listTxFunc(prefix, &kvs, &found, &sxexec.ExecutionOptions{})(tx); err != nil {
		return "", err
	}

	execs, err := bs.unmarshalExecutions(kvs, nil)
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
		status = JobStatusSuccess
	} else if failed > 0 && success == 0 {
		status = JobStatusFailed
	} else if failed > 0 && success > 0 {
		status = JobStatusPartiallyFailed
	}

	return status, nil
}

func (bs *BuntdbStore) jobHasMetadata(job *Job, metadata map[string]string) bool {
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
