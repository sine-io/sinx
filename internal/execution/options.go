package execution

import "time"

// ExecutionOptions additional options like "Sort" will be ready for JSON marshall
type ExecutionOptions struct {
	Sort     string
	Order    string
	Timezone *time.Location
}
