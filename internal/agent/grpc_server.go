package agent

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

var (
	// ErrExecutionDoneForDeletedJob is returned when an execution done
	// is received for a non existent job.
	ErrExecutionDoneForDeletedJob = errors.New("grpc: Received execution done for a deleted job")
	// ErrRPCDialing is returned on dialing fail.
	ErrRPCDialing = errors.New("grpc: Error dialing, verify the network connection to the server")
	// ErrNotLeader is the error returned when the operation need the node to be the leader,
	// but the current node is not the leader.
	ErrNotLeader = errors.New("grpc: Error, server is not leader, this operation should be run on the leader")
	// ErrBrokenStream is the error that indicates a sudden disconnection of the agent streaming an execution
	ErrBrokenStream = errors.New("grpc: Error on execution streaming, agent connection was abruptly terminated")
)

// GRPCServer is the local implementation of the gRPC server interface.
type GRPCServer struct {
	sxproto.SinxServer
	agent *Agent

	logger zerolog.Logger
}

// NewGRPCServer creates and returns an instance of a SinxGRPCServer implementation
func NewGRPCServer(agent *Agent) *GRPCServer {
	return &GRPCServer{
		agent: agent,

		logger: zerolog.New(zerolog.NewConsoleWriter()),
	}
}

func (grpcs *GRPCServer) WithLogger(logger *zerolog.Logger) *GRPCServer {
	grpcs.logger = logger.Hook()

	return grpcs
}

// Serve creates and start a new gRPC server
func (grpcs *GRPCServer) Serve(lis net.Listener) error {
	grpcServer := grpc.NewServer()
	sxproto.RegisterSinxServer(grpcServer, grpcs)

	as := NewGRPCAgentServer(grpcs.agent)
	sxproto.RegisterAgentServer(grpcServer, as)
	go grpcServer.Serve(lis)

	return nil
}

// SetJob broadcast a state change to the cluster members that will store the job.
// Then restart the scheduler
// This only works on the leader
func (grpcs *GRPCServer) SetJob(ctx context.Context, setJobReq *sxproto.SetJobRequest) (*sxproto.SetJobResponse, error) {
	defer metrics.MeasureSince([]string{"grpc", "set_job"}, time.Now())

	grpcs.logger.Debug().
		Str("job", setJobReq.Job.Name).
		Msg("grpc: Received SetJob")

	if err := grpcs.agent.applySetJob(setJobReq.Job); err != nil {
		return nil, err
	}

	// If everything is ok, add the job to the scheduler
	job := NewJobFromProto(setJobReq.Job)
	job.Agent = grpcs.agent
	if err := grpcs.agent.sched.AddJob(job); err != nil {
		return nil, err
	}

	return &sxproto.SetJobResponse{}, nil
}

// DeleteJob broadcast a state change to the cluster members that will delete the job.
// This only works on the leader
func (grpcs *GRPCServer) DeleteJob(ctx context.Context, delJobReq *sxproto.DeleteJobRequest) (*sxproto.DeleteJobResponse, error) {
	defer metrics.MeasureSince([]string{"grpc", "delete_job"}, time.Now())

	grpcs.logger.Debug().Str("job", delJobReq.GetJobName()).Msg("grpc: Received DeleteJob")

	cmd, err := Encode(DeleteJobType, delJobReq)
	if err != nil {
		return nil, err
	}
	af := grpcs.agent.raft.Apply(cmd, raftTimeout)
	if err := af.Error(); err != nil {
		return nil, err
	}
	res := af.Response()
	job, ok := res.(*Job)
	if !ok {
		return nil, fmt.Errorf("grpc: Error wrong response from apply in DeleteJob: %v", res)
	}
	jpb := job.ToProto()

	// If everything is ok, remove the job
	grpcs.agent.sched.RemoveJob(job.Name)
	if job.Ephemeral {
		grpcs.logger.Info().Str("job", job.Name).Msg("grpc: Done deleting ephemeral job")
	}

	return &sxproto.DeleteJobResponse{Job: jpb}, nil
}

// GetJob loads the job from the datastore
func (grpcs *GRPCServer) GetJob(ctx context.Context, getJobReq *sxproto.GetJobRequest) (*sxproto.GetJobResponse, error) {
	defer metrics.MeasureSince([]string{"grpc", "get_job"}, time.Now())
	grpcs.logger.Debug().Str("job", getJobReq.JobName).Msg("grpc: Received GetJob")

	j, err := grpcs.agent.Store.GetJob(getJobReq.JobName, nil)
	if err != nil {
		return nil, err
	}

	gjr := &sxproto.GetJobResponse{
		Job: &sxproto.Job{},
	}

	// Copy the data structure
	gjr.Job.Name = j.Name
	gjr.Job.Executor = j.Executor
	gjr.Job.ExecutorConfig = j.ExecutorConfig

	return gjr, nil
}

// ExecutionDone saves the execution to the store
func (grpcs *GRPCServer) ExecutionDone(ctx context.Context, execDoneReq *sxproto.ExecutionDoneRequest) (*sxproto.ExecutionDoneResponse, error) {
	defer metrics.MeasureSince([]string{"grpc", "execution_done"}, time.Now())
	grpcs.logger.Debug().
		Int64("group", execDoneReq.Execution.Group).
		Str("job", execDoneReq.Execution.JobName).
		Str("from", execDoneReq.Execution.NodeName).
		Msg("grpc: Received execution done")

	// Get the leader address and compare with the current node address.
	// Forward the request to the leader in case current node is not the leader.
	if !grpcs.agent.IsLeader() {
		addr := grpcs.agent.raft.Leader()
		grpcs.agent.GRPCClient.ExecutionDone(string(addr), NewExecutionFromProto(execDoneReq.Execution))
		return nil, ErrNotLeader
	}

	// This is the leader at this point, so process the execution, encode the value and apply the log to the cluster.
	// Get the defined output types for the job, and call them
	job, err := grpcs.agent.Store.GetJob(execDoneReq.Execution.JobName, nil)
	if err != nil {
		return nil, err
	}

	pbex := *execDoneReq.Execution
	for k, v := range job.Processors {
		grpcs.logger.Info().Str("plugin", k).Msg("grpc: Processing execution with plugin")
		if processor, ok := grpcs.agent.ProcessorPlugins[k]; ok {
			v["reporting_node"] = grpcs.agent.config.NodeName
			pbex = processor.Process(&sxplugin.ProcessorArgs{Execution: pbex, Config: v})
		} else {
			grpcs.logger.Error().Str("plugin", k).Msg("grpc: Specified plugin not found")
		}
	}

	execDoneReq.Execution = &pbex
	cmd, err := Encode(ExecutionDoneType, execDoneReq)
	if err != nil {
		return nil, err
	}
	af := grpcs.agent.raft.Apply(cmd, raftTimeout)
	if err := af.Error(); err != nil {
		return nil, err
	}

	// Retrieve the fresh, updated job from the store to work on stored values
	job, err = grpcs.agent.Store.GetJob(job.Name, nil)
	if err != nil {
		grpcs.logger.Error().
			Err(err).
			Str("job", execDoneReq.Execution.JobName).
			Msg("grpc: Error retrieving job from store")

		return nil, err
	}

	// If the execution failed, retry it until retries limit (default: don't retry)
	// Don't retry if the status is unknown
	execution := NewExecutionFromProto(&pbex)
	if !execution.Success &&
		uint(execution.Attempt) < job.Retries+1 &&
		!strings.HasPrefix(execution.Output, ErrBrokenStream.Error()) {
		// Increment the attempt counter
		execution.Attempt++

		// Keep all execution properties intact except the last output
		execution.Output = ""

		eb := execution.CalculateExponentialBackoff()
		grpcs.logger.Debug().
			Uint("attempt", execution.Attempt).
			Any("execution", execution).
			Any("backoff", eb).
			Msg("grpc: Retrying execution")

		time.Sleep(eb)

		if _, err := grpcs.agent.Run(job.Name, execution); err != nil {
			return nil, err
		}

		return &sxproto.ExecutionDoneResponse{
			From:    grpcs.agent.config.NodeName,
			Payload: []byte("retry"),
		}, nil
	}

	exg, err := grpcs.agent.Store.GetExecutionGroup(execution,
		&ExecutionOptions{
			Timezone: job.GetTimeLocation(),
		},
	)
	if err != nil {
		grpcs.logger.Error().
			Err(err).
			Int64("group", execution.Group).
			Msg("grpc: Error getting execution group.")

		return nil, err
	}

	// Send notification
	if err := SendPostNotifications(grpcs.agent.config, execution, exg, job, grpcs.logger); err != nil {
		return nil, err
	}

	// Jobs that have dependent jobs are a bit more expensive because we need to call the Status() method for every execution.
	// Check first if there's dependent jobs and then check for the job status to begin execution dependent jobs on success.
	if len(job.DependentJobs) > 0 && job.Status == StatusSuccess {
		for _, djn := range job.DependentJobs {
			dj, err := grpcs.agent.Store.GetJob(djn, nil)
			if err != nil {
				return nil, err
			}
			dj.Agent = grpcs.agent
			grpcs.logger.Debug().Str("job", djn).Msg("grpc: Running dependent job")
			dj.Run()
		}
	}

	if job.Ephemeral && job.Status == StatusSuccess {
		if _, err := grpcs.DeleteJob(ctx, &sxproto.DeleteJobRequest{JobName: job.Name}); err != nil {
			return nil, err
		}
		return &sxproto.ExecutionDoneResponse{
			From:    grpcs.agent.config.NodeName,
			Payload: []byte("deleted"),
		}, nil
	}

	return &sxproto.ExecutionDoneResponse{
		From:    grpcs.agent.config.NodeName,
		Payload: []byte("saved"),
	}, nil
}

// Leave calls the Stop method, stopping everything in the server
func (grpcs *GRPCServer) Leave(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return in, grpcs.agent.Stop()
}

// RunJob runs a job in the cluster
func (grpcs *GRPCServer) RunJob(ctx context.Context, req *sxproto.RunJobRequest) (*sxproto.RunJobResponse, error) {
	ex := NewExecution(req.JobName)
	job, err := grpcs.agent.Run(req.JobName, ex)
	if err != nil {
		return nil, err
	}
	jpb := job.ToProto()

	return &sxproto.RunJobResponse{Job: jpb}, nil
}

// ToggleJob toggle the enablement of a job
func (grpcs *GRPCServer) ToggleJob(ctx context.Context, getJobReq *sxproto.ToggleJobRequest) (*sxproto.ToggleJobResponse, error) {
	return nil, nil
}

// RaftGetConfiguration get raft config
func (grpcs *GRPCServer) RaftGetConfiguration(ctx context.Context, in *emptypb.Empty) (*sxproto.RaftGetConfigurationResponse, error) {
	// We can't fetch the leader and the configuration atomically with
	// the current Raft API.
	future := grpcs.agent.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}

	// Index the information about the servers.
	serverMap := make(map[raft.ServerAddress]serf.Member)
	for _, member := range grpcs.agent.serf.Members() {
		valid, parts := isServer(member)
		if !valid {
			continue
		}

		addr := (&net.TCPAddr{IP: member.Addr, Port: parts.Port}).String()
		serverMap[raft.ServerAddress(addr)] = member
	}

	// Fill out the reply.
	leader := grpcs.agent.raft.Leader()
	reply := &sxproto.RaftGetConfigurationResponse{}
	reply.Index = future.Index()
	for _, server := range future.Configuration().Servers {
		node := "(unknown)"
		raftProtocolVersion := "unknown"
		if member, ok := serverMap[server.Address]; ok {
			node = member.Name
			if raftVsn, ok := member.Tags["raft_vsn"]; ok {
				raftProtocolVersion = raftVsn
			}
		}

		entry := &sxproto.RaftServer{
			Id:           string(server.ID),
			Node:         node,
			Address:      string(server.Address),
			Leader:       server.Address == leader,
			Voter:        server.Suffrage == raft.Voter,
			RaftProtocol: raftProtocolVersion,
		}
		reply.Servers = append(reply.Servers, entry)
	}
	return reply, nil
}

// RaftRemovePeerByID is used to kick a stale peer (one that is in the Raft
// quorum but no longer known to Serf or the catalog) by address in the form of
// "IP:port". The reply argument is not used, but is required to fulfill the RPC
// interface.
func (grpcs *GRPCServer) RaftRemovePeerByID(ctx context.Context, in *sxproto.RaftRemovePeerByIDRequest) (*emptypb.Empty, error) {
	// Since this is an operation designed for humans to use, we will return
	// an error if the supplied id isn't among the peers since it's
	// likely they screwed up.
	{
		future := grpcs.agent.raft.GetConfiguration()
		if err := future.Error(); err != nil {
			return nil, err
		}
		for _, s := range future.Configuration().Servers {
			if s.ID == raft.ServerID(in.Id) {
				goto REMOVE
			}
		}
		return nil, fmt.Errorf("id %q was not found in the Raft configuration", in.Id)
	}

REMOVE:
	// The Raft library itself will prevent various forms of foot-shooting,
	// like making a configuration with no voters. Some consideration was
	// given here to adding more checks, but it was decided to make this as
	// low-level and direct as possible. We've got ACL coverage to lock this
	// down, and if you are an operator, it's assumed you know what you are
	// doing if you are calling this. If you remove a peer that's known to
	// Serf, for example, it will come back when the leader does a reconcile
	// pass.
	future := grpcs.agent.raft.RemoveServer(raft.ServerID(in.Id), 0, 0)
	if err := future.Error(); err != nil {
		grpcs.logger.Warn().Err(err).Str("peer", in.Id).Msg("failed to remove Raft peer")
		return nil, err
	}

	grpcs.logger.Warn().Str("peer", in.Id).Msg("removed Raft peer")

	return new(emptypb.Empty), nil
}

// GetActiveExecutions returns the active executions on the server node
func (grpcs *GRPCServer) GetActiveExecutions(ctx context.Context, in *emptypb.Empty) (*sxproto.GetActiveExecutionsResponse, error) {
	defer metrics.MeasureSince([]string{"grpc", "agent_run"}, time.Now())

	var executions []*sxproto.Execution
	grpcs.agent.activeExecutions.Range(func(k, v interface{}) bool {
		e := v.(*sxproto.Execution)
		executions = append(executions, e)
		return true
	})

	return &sxproto.GetActiveExecutionsResponse{
		Executions: executions,
	}, nil
}

// SetExecution broadcast a state change to the cluster members that will store the execution.
// This only works on the leader
func (grpcs *GRPCServer) SetExecution(ctx context.Context, execution *sxproto.Execution) (*emptypb.Empty, error) {
	defer metrics.MeasureSince([]string{"grpc", "set_execution"}, time.Now())
	grpcs.logger.Debug().
		Str("execution", execution.Key()).
		Msg("grpc: Received SetExecution")

	cmd, err := Encode(SetExecutionType, execution)
	if err != nil {
		grpcs.logger.Fatal().Err(err).Msg("agent: encode error in SetExecution")
		return nil, err
	}
	af := grpcs.agent.raft.Apply(cmd, raftTimeout)
	if err := af.Error(); err != nil {
		grpcs.logger.Fatal().Err(err).Msg("agent: error applying SetExecutionType")
		return nil, err
	}

	return new(emptypb.Empty), nil
}

// Encode is used to encode a Protoc object with type prefix
func Encode(t MessageType, msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(uint8(t))
	m, err := proto.Marshal(msg.(proto.Message))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(m)
	return buf.Bytes(), err
}
