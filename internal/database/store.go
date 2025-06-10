package database

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/tidwall/buntdb"
)

const (
	// MaxExecutions to maintain in the storage
	MaxExecutions = 100

	jobsPrefix       = "jobs"
	executionsPrefix = "executions"
)

type kv struct {
	Key   string
	Value []byte
}

// NewBuntJobDB creates a new NewBuntJobDB instance.
func NewBuntJobDB(logger zerolog.Logger) (*BuntJobDB, error) {
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

	store := &BuntJobDB{
		db:     db,
		lock:   &sync.Mutex{},
		logger: logger,
	}

	return store, nil
}

// DB is the getter for the BuntDB instance
func (bjd *BuntJobDB) DB() *buntdb.DB {
	return bjd.db
}
