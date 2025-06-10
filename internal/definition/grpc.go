package definition

import (
	"net"

	"google.golang.org/grpc"

	sxexec "github.com/sine-io/sinx/internal/execution"
	sxjob "github.com/sine-io/sinx/internal/job"
	sxproto "github.com/sine-io/sinx/types"
)

// DkronGRPCServer defines the basics that a gRPC server should implement.
type DkronGRPCServer interface {
	sxproto.DkronServer
	Serve(net.Listener) error
}

// DkronGRPCClient defines the interface that any gRPC client for
// dkron should implement.
type DkronGRPCClient interface {
	Connect(string) (*grpc.ClientConn, error)
	ExecutionDone(string, *sxexec.Execution) error
	GetJob(string, string) (*sxjob.Job, error)
	SetJob(*sxjob.Job) error
	DeleteJob(string) (*sxjob.Job, error)
	Leave(string) error
	RunJob(string) (*sxjob.Job, error)
	RaftGetConfiguration(string) (*sxproto.RaftGetConfigurationResponse, error)
	RaftRemovePeerByID(string, string) error
	GetActiveExecutions(string) ([]*sxproto.Execution, error)
	SetExecution(execution *sxproto.Execution) error
	AgentRun(addr string, job *sxproto.Job, execution *sxproto.Execution) error
}
