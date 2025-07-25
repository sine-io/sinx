package agent

import (
	"errors"
	"time"

	"github.com/armon/circbuf"
	metrics "github.com/armon/go-metrics"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"

	sxexec "github.com/sine-io/sinx/internal/execution"
	sxproto "github.com/sine-io/sinx/types"
)

const (
	// maxBufSize limits how much data we collect from a handler.
	maxBufSize = 256000
)

type statusAgentHelper struct {
	execution *sxproto.Execution
	stream    sxproto.Agent_AgentRunServer
}

func (s *statusAgentHelper) Update(b []byte, c bool) (int64, error) {
	s.execution.Output = b
	// Send partial execution
	if err := s.stream.Send(&sxproto.AgentRunStream{
		Execution: s.execution,
	}); err != nil {
		return 0, err
	}
	return 0, nil
}

// GRPCAgentServer is the local implementation of the gRPC server interface.
type GRPCAgentServer struct {
	sxproto.AgentServer
	agent *Agent

	logger zerolog.Logger
}

// NewGRPCAgentServer creates and returns an instance of a AgentServer implementation
func NewGRPCAgentServer(agent *Agent) sxproto.AgentServer {
	return &GRPCAgentServer{
		agent:  agent,
		logger: zerolog.New(zerolog.NewConsoleWriter()),
	}
}

func (as *GRPCAgentServer) WithLogger(logger *zerolog.Logger) *GRPCAgentServer {
	as.logger = logger.Hook()

	return as
}

// AgentRun is called when an agent starts running a job and lasts all execution,
// the agent will stream execution progress to the server.
func (as *GRPCAgentServer) AgentRun(req *sxproto.AgentRunRequest, stream sxproto.Agent_AgentRunServer) error {
	defer metrics.MeasureSince([]string{"grpc_agent", "agent_run"}, time.Now())

	job := req.Job
	execution := req.Execution

	as.logger.Info().Str("job", job.Name).Msg("grpc_agent: Starting job")

	output, _ := circbuf.NewBuffer(maxBufSize)

	var success bool

	jex := job.Executor
	exc := job.ExecutorConfig

	// Send the first update with the initial execution state to be stored in the server
	execution.StartedAt = timestamppb.Now()
	execution.NodeName = as.agent.config.NodeName

	if err := stream.Send(&sxproto.AgentRunStream{
		Execution: execution,
	}); err != nil {
		return err
	}

	if jex == "" {
		return errors.New("grpc_agent: No executor defined, nothing to do")
	}

	// Check if executor exists
	if executor, ok := as.agent.ExecutorPlugins[jex]; ok {
		as.logger.Debug().Str("plugin", jex).Msg("grpc_agent: calling executor plugin")

		runningExecutions.Store(execution.GetGroup(), execution)
		out, err := executor.Execute(&sxproto.ExecuteRequest{
			JobName: job.Name,
			Config:  exc,
		}, &statusAgentHelper{
			stream:    stream,
			execution: execution,
		})

		if err == nil && out.Error != "" {
			err = errors.New(out.Error)
		}
		if err != nil {
			as.logger.Error().Err(err).Str("job", job.Name).Any("plugin", executor).Msg(
				"grpc_agent: command error output")

			success = false
			_, _ = output.Write([]byte(err.Error() + "\n"))
		} else {
			success = true
		}

		if out != nil {
			_, _ = output.Write(out.Output)
		}
	} else {
		as.logger.Error().Str("executor", jex).Msg("grpc_agent: Specified executor is not present")
		_, _ = output.Write([]byte("grpc_agent: Specified executor is not present"))
	}

	execution.FinishedAt = timestamppb.Now()
	execution.Success = success
	execution.Output = output.Bytes()

	runningExecutions.Delete(execution.GetGroup())

	// Send the final execution
	if err := stream.Send(&sxproto.AgentRunStream{
		Execution: execution,
	}); err != nil {
		// In case of error means that maybe the server is gone so fallback to ExecutionDone
		as.logger.Error().Err(err).Str("job", job.Name).Msg(
			"grpc_agent: error sending the final execution, falling back to ExecutionDone")

		rpcServer, err := as.agent.CheckAndSelectServer()
		if err != nil {
			return err
		}
		return as.agent.GRPCClient.ExecutionDone(rpcServer, sxexec.NewExecutionFromProto(execution))
	}

	return nil
}
