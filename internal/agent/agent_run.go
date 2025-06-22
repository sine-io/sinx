package agent

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/serf/serf"

	sxexec "github.com/sine-io/sinx/internal/execution"
)

// Run call the agents to run a job. Returns a job with its new status and next schedule.
func (a *Agent) RunAgent(jobName string, ex *sxexec.Execution) (*Job, error) {
	job, err := a.Storage.GetJob(jobName, nil)
	if err != nil {
		return nil, fmt.Errorf("agent: Run error retrieving job: %s from store: %w", jobName, err)
	}

	// In case the job is not a child job, compute the next execution time
	if job.ParentJob == "" {
		if ej, ok := a.sched.GetCronEntryJob(jobName); ok {
			job.Next = ej.Entry.Next
			if err := a.applySetJob(job.ToProto()); err != nil {
				return nil, fmt.Errorf("agent: Run error storing job %s before running: %w", jobName, err)
			}
		} else {
			return nil, fmt.Errorf("agent: Run error retrieving job: %s from scheduler", jobName)
		}
	}

	// In the first execution attempt we build and filter the target nodes
	// but we use the existing node target in case of retry.
	var targetNodes []Node
	if ex.Attempt <= 1 {
		targetNodes = a.getTargetNodes(job.Tags, defaultSelector)
	} else {
		// In case of retrying, find the node or return with an error
		for _, m := range a.serf.Members() {
			if ex.NodeName == m.Name {
				if m.Status == serf.StatusAlive {
					targetNodes = []Node{m}
					break
				} else {
					return nil, fmt.Errorf("retry node is gone: %s for job %s", ex.NodeName, ex.JobName)
				}
			}
		}
	}

	// In case no nodes found, return reporting the error
	if len(targetNodes) < 1 {
		return nil, fmt.Errorf("no target nodes found to run job %s", ex.JobName)
	}
	a.logger.Debug().Any("nodes", targetNodes).Msg("agent: Filtered nodes to run")

	var wg sync.WaitGroup
	for _, v := range targetNodes {
		// Determine node address
		addr, ok := v.Tags["rpc_addr"]
		if !ok {
			addr = v.Addr.String()
		}

		// Call here client GRPC AgentRun
		wg.Add(1)
		go func(node string, wg *sync.WaitGroup) {
			defer wg.Done()

			a.logger.Info().Str("jog_name", job.Name).Str("node", node).Msg("agent: Calling AgentRun")

			err := a.GRPCClient.AgentRun(node, job.ToProto(), ex.ToProto())
			if err != nil {
				a.logger.Error().Str("job_name", job.Name).Str("node", node).Err(err).Msg("agent: Error calling AgentRun")
			}
		}(addr, &wg)
	}

	wg.Wait()

	return job, nil
}

func (a *Agent) getTargetNodes(tags map[string]string, selectFunc func([]Node) int) []Node {
	bareTags, cardinality := a.cleanTags(tags)
	nodes := a.getQualifyingNodes(a.serf.Members(), bareTags)

	return selectNodes(nodes, cardinality, selectFunc)
}

// cleanTags takes the tag spec and returns strictly key:value pairs
// along with the lowest cardinality specified
func (a *Agent) cleanTags(tags map[string]string) (map[string]string, int) {
	cardinality := int(^uint(0) >> 1) // MaxInt

	cleanTags := make(map[string]string, len(tags))

	for k, v := range tags {
		vparts := strings.Split(v, ":")

		cleanTags[k] = vparts[0]

		// If a cardinality is specified (i.e. "value:3") and it is lower than our
		// max cardinality, lower the max
		if len(vparts) == 2 {
			tagCard, err := strconv.Atoi(vparts[1])
			if err != nil {
				// Tag value is malformed
				tagCard = 0
				a.logger.Error().Msgf("improper cardinality specified for tag %s: %v", k, vparts[1])
			}

			if tagCard < cardinality {
				cardinality = tagCard
			}
		}
	}

	return cleanTags, cardinality
}

// getQualifyingNodes returns all nodes in the cluster that are
// alive, in this agent's region and have all given tags
func (a *Agent) getQualifyingNodes(nodes []Node, bareTags map[string]string) []Node {
	// Determine the usable set of nodes
	qualifiers := filterArray(nodes, func(node Node) bool {
		return node.Status == serf.StatusAlive &&
			node.Tags["region"] == a.config.Region &&
			nodeMatchesTags(node, bareTags)
	})

	return qualifiers
}

// Returns all items from an array for which filterFunc returns true,
func filterArray(arr []Node, filterFunc func(Node) bool) []Node {
	for i := len(arr) - 1; i >= 0; i-- {
		if !filterFunc(arr[i]) {
			arr[i] = arr[len(arr)-1]
			arr = arr[:len(arr)-1]
		}
	}
	return arr
}

// nodeMatchesTags tests if a node matches all of the provided tags
func nodeMatchesTags(node serf.Member, tags map[string]string) bool {
	for k, v := range tags {
		nodeVal, present := node.Tags[k]
		if !present {
			return false
		}
		if nodeVal != v {
			return false
		}
	}
	// If we matched all key:value pairs, the node matches the tags
	return true
}

// selectNodes selects at most #cardinality from the given nodes using the selectFunc
func selectNodes(nodes []Node, cardinality int, selectFunc func([]Node) int) []Node {
	// Return all nodes immediately if they're all going to be selected
	numNodes := len(nodes)
	if numNodes <= cardinality {
		return nodes
	}

	for ; cardinality > 0; cardinality-- {
		// Select a node
		chosenIndex := selectFunc(nodes[:numNodes])

		// Swap picked node with the last one and reduce choices so it can't get picked again
		nodes[numNodes-1], nodes[chosenIndex] = nodes[chosenIndex], nodes[numNodes-1]
		numNodes--
	}

	return nodes[numNodes:]
}

// The default selector function for getTargetNodes/selectNodes
func defaultSelector(nodes []Node) int {
	return rand.IntN(len(nodes))
}
