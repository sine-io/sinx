package agent

import (
	"fmt"
	"io"
	"time"

	metrics "github.com/armon/go-metrics"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sxproto "github.com/sine-io/sinx/types"
)

// DkronGRPCClient defines the interface that any gRPC client for
// dkron should implement.
type DkronGRPCClient interface {
	Connect(string) (*grpc.ClientConn, error)
	ExecutionDone(string, *Execution) error
	GetJob(string, string) (*Job, error)
	SetJob(*Job) error
	DeleteJob(string) (*Job, error)
	Leave(string) error
	RunJob(string) (*Job, error)
	RaftGetConfiguration(string) (*sxproto.RaftGetConfigurationResponse, error)
	RaftRemovePeerByID(string, string) error
	GetActiveExecutions(string) ([]*sxproto.Execution, error)
	SetExecution(execution *sxproto.Execution) error
	AgentRun(addr string, job *sxproto.Job, execution *sxproto.Execution) error
}

// GRPCClient is the local implementation of the DkronGRPCClient interface.
type GRPCClient struct {
	dialOpt []grpc.DialOption
	agent   *Agent
}

// NewGRPCClient returns a new instance of the gRPC client.
func NewGRPCClient(dialOpt grpc.DialOption, agent *Agent) DkronGRPCClient {
	if dialOpt == nil {
		dialOpt = grpc.WithInsecure()
	}
	return &GRPCClient{
		dialOpt: []grpc.DialOption{
			dialOpt,
			grpc.WithBlock(),
		},
		agent: agent,
	}
}

// Connect dialing to a gRPC server
func (grpcc *GRPCClient) Connect(addr string) (*grpc.ClientConn, error) {
	// Initiate a connection with the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpcc.dialOpt...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ExecutionDone calls the ExecutionDone gRPC method
func (grpcc *GRPCClient) ExecutionDone(addr string, execution *Execution) error {
	defer metrics.MeasureSince([]string{"grpc", "call_execution_done"}, time.Now())
	var conn *grpc.ClientConn

	conn, err := grpcc.Connect(addr)
	if err != nil {
		grpcc.agent.logger.Err(err).
			Str("method", "ExecutionDone").
			Str("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	d := sxproto.NewDkronClient(conn)
	edr, err := d.ExecutionDone(context.Background(), &sxproto.ExecutionDoneRequest{Execution: execution.ToProto()})
	if err != nil {
		if err.Error() == fmt.Sprintf("rpc error: code = Unknown desc = %s", ErrNotLeader.Error()) {
			grpcc.agent.logger.Info().Msg("grpc: ExecutionDone forwarded to the leader")
			return nil
		}

		grpcc.agent.logger.Error().Err(err).
			Str("method", "ExecutionDone").
			Str("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return err
	}

	grpcc.agent.logger.Debug().
		Str("method", "ExecutionDone").
		Str("server_addr", addr).
		Str("from", edr.From).
		Str("payload", string(edr.Payload)).
		Msg("grpc: Response from method")

	return nil
}

// GetJob calls GetJob gRPC method in the server
func (grpcc *GRPCClient) GetJob(addr, jobName string) (*Job, error) {
	defer metrics.MeasureSince([]string{"grpc", "get_job"}, time.Now())
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(addr)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "GetJob").
			Str("server_addr", addr).
			Msg("grpc: error dialing.")

		return nil, err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	gjr, err := d.GetJob(context.Background(), &sxproto.GetJobRequest{JobName: jobName})
	if err != nil {
		grpcc.agent.logger.Error().Err(err).
			Str("method", "GetJob").
			Str("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return nil, err
	}

	return NewJobFromProto(gjr.Job), nil
}

// Leave calls Leave method on the gRPC server
func (grpcc *GRPCClient) Leave(addr string) error {
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(addr)
	if err != nil {
		grpcc.agent.logger.Error().Err(err).
			Str("method", "Leave").
			Str("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	_, err = d.Leave(context.Background(), &emptypb.Empty{})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "Leave").
			Str("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return err
	}

	return nil
}

// SetJob calls the leader passing the job
func (grpcc *GRPCClient) SetJob(job *Job) error {
	var conn *grpc.ClientConn

	addr := grpcc.agent.raft.Leader()

	// Initiate a connection with the server
	conn, err := grpcc.Connect(string(addr))
	if err != nil {
		grpcc.agent.logger.Error().Err(err).
			Str("method", "SetJob").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	_, err = d.SetJob(context.Background(), &sxproto.SetJobRequest{
		Job: job.ToProto(),
	})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "SetJob").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return err
	}

	return nil
}

// DeleteJob calls the leader passing the job name
func (grpcc *GRPCClient) DeleteJob(jobName string) (*Job, error) {
	var conn *grpc.ClientConn

	addr := grpcc.agent.raft.Leader()

	// Initiate a connection with the server
	conn, err := grpcc.Connect(string(addr))
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "DeleteJob").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return nil, err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	res, err := d.DeleteJob(context.Background(), &sxproto.DeleteJobRequest{
		JobName: jobName,
	})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "DeleteJob").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return nil, err
	}

	job := NewJobFromProto(res.Job)

	return job, nil
}

// RunJob calls the leader passing the job name
func (grpcc *GRPCClient) RunJob(jobName string) (*Job, error) {
	var conn *grpc.ClientConn

	addr := grpcc.agent.raft.Leader()

	// Initiate a connection with the server
	conn, err := grpcc.Connect(string(addr))
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RunJob").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return nil, err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	res, err := d.RunJob(context.Background(), &sxproto.RunJobRequest{
		JobName: jobName,
	})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RunJob").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return nil, err
	}

	job := NewJobFromProto(res.Job)

	return job, nil
}

// RaftGetConfiguration get the current raft configuration of peers
func (grpcc *GRPCClient) RaftGetConfiguration(addr string) (*sxproto.RaftGetConfigurationResponse, error) {
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(addr)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RaftGetConfiguration").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return nil, err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	res, err := d.RaftGetConfiguration(context.Background(), &emptypb.Empty{})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RaftGetConfiguration").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return nil, err
	}

	return res, nil
}

// RaftRemovePeerByID remove a raft peer
func (grpcc *GRPCClient) RaftRemovePeerByID(addr, peerID string) error {
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(addr)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RaftRemovePeerByID").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	_, err = d.RaftRemovePeerByID(context.Background(),
		&sxproto.RaftRemovePeerByIDRequest{Id: peerID},
	)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "RaftRemovePeerByID").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return err
	}

	return nil
}

// GetActiveExecutions returns the active executions of a server node
func (grpcc *GRPCClient) GetActiveExecutions(addr string) ([]*sxproto.Execution, error) {
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(addr)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "GetActiveExecutions").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return nil, err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	gaer, err := d.GetActiveExecutions(context.Background(), &emptypb.Empty{})
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "GetActiveExecutions").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return nil, err
	}

	return gaer.Executions, nil
}

// SetExecution calls the leader passing the execution
func (grpcc *GRPCClient) SetExecution(execution *sxproto.Execution) error {
	var conn *grpc.ClientConn

	addr := grpcc.agent.raft.Leader()

	// Initiate a connection with the server
	conn, err := grpcc.Connect(string(addr))
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "SetExecution").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	// Synchronous call
	d := sxproto.NewDkronClient(conn)
	_, err = d.SetExecution(context.Background(), execution)
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "SetExecution").
			Any("server_addr", addr).
			Msg("grpc: Error calling gRPC method")

		return err
	}
	return nil
}

// AgentRun runs a job in the given agent
func (grpcc *GRPCClient) AgentRun(addr string, job *sxproto.Job, execution *sxproto.Execution) error {
	defer metrics.MeasureSince([]string{"grpc_client", "agent_run"}, time.Now())
	var conn *grpc.ClientConn

	// Initiate a connection with the server
	conn, err := grpcc.Connect(string(addr))
	if err != nil {

		grpcc.agent.logger.Error().Err(err).
			Str("method", "AgentRun").
			Any("server_addr", addr).
			Msg("grpc: error dialing.")

		return err
	}
	defer conn.Close()

	// Streaming call
	a := sxproto.NewAgentClient(conn)
	stream, err := a.AgentRun(context.Background(), &sxproto.AgentRunRequest{
		Job:       job,
		Execution: execution,
	})
	if err != nil {
		return err
	}

	var first bool
	for {
		ars, err := stream.Recv()

		// Stream ends
		if err == io.EOF {
			addr := grpcc.agent.raft.Leader()
			if err := grpcc.ExecutionDone(string(addr), NewExecutionFromProto(execution)); err != nil {
				return err
			}
			return nil
		}

		// Error received from the stream
		if err != nil {
			// At this point the execution status will be unknown, set the FinishedAt time and an explanatory message
			execution.FinishedAt = timestamppb.Now()
			execution.Output = []byte(ErrBrokenStream.Error() + ": " + err.Error())

			grpcc.agent.logger.Error().Err(err).Err(ErrBrokenStream).Send()

			addr := grpcc.agent.raft.Leader()
			if err := grpcc.ExecutionDone(string(addr), NewExecutionFromProto(execution)); err != nil {
				return err
			}
			return err
		}

		// Registers an active stream
		grpcc.agent.activeExecutions.Store(ars.Execution.Key(), ars.Execution)

		grpcc.agent.logger.Debug().
			Str("key", ars.Execution.Key()).
			Msg("grpc: received execution stream")

		execution = ars.Execution
		defer grpcc.agent.activeExecutions.Delete(execution.Key())

		// Store the received execution in the raft log and store
		if !first {
			if err := grpcc.SetExecution(ars.Execution); err != nil {
				return err
			}
			first = true
		}

		// Notify the starting of the execution
		if err := SendPreNotifications(grpcc.agent.config, NewExecutionFromProto(execution), nil, NewJobFromProto(job), grpcc.agent.logger); err != nil {
			grpcc.agent.logger.Error().
				Err(err).
				Str("job_name", job.Name).
				Str("node", grpcc.agent.config.NodeName).
				Msg("agent: Error sending start notification")
		}
	}
}
