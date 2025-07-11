package plugin

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	sxproto "github.com/sine-io/sinx/types"
)

type StatusHelper interface {
	Update([]byte, bool) (int64, error)
}

// Executor is the interface that we're exposing as a goplugin.
type Executor interface {
	Execute(args *sxproto.ExecuteRequest, cb StatusHelper) (*sxproto.ExecuteResponse, error)
}

// ExecutorPluginConfig is the plugin config
type ExecutorPluginConfig map[string]string

// This is the implementation of goplugin.Plugin so we can serve/consume this.
// We also implement GRPCPlugin so that this plugin can be served over
// gRPC.
type ExecutorPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
	Executor Executor
}

func (p *ExecutorPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	sxproto.RegisterExecutorServer(s, ExecutorServer{Impl: p.Executor, broker: broker})
	return nil
}

func (p *ExecutorPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ExecutorClient{client: sxproto.NewExecutorClient(c), broker: broker}, nil
}

type Broker interface {
	NextId() uint32
	AcceptAndServe(id uint32, s func([]grpc.ServerOption) *grpc.Server)
}

// Here is the gRPC client that GRPCClient talks to.
type ExecutorClient struct {
	// This is the real implementation
	client sxproto.ExecutorClient
	broker Broker
}

func (m *ExecutorClient) Execute(args *sxproto.ExecuteRequest, cb StatusHelper) (*sxproto.ExecuteResponse, error) {
	// This is where the magic conversion to Proto happens
	statusHelperServer := &GRPCStatusHelperServer{Impl: cb}

	initChan := make(chan bool, 1)
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		sxproto.RegisterStatusHelperServer(s, statusHelperServer)
		initChan <- true

		return s
	}

	brokerID := m.broker.NextId()
	go func() {
		m.broker.AcceptAndServe(brokerID, serverFunc)
		// AcceptAndServe might terminate without calling serverFunc
		// To prevent eternal blocking, send 'init done' signal
		initChan <- true
	}()

	// Wait for s to be initialized in the goroutine
	<-initChan

	args.StatusServer = brokerID
	r, err := m.client.Execute(context.Background(), args)

	/* In some cases the server cannot start (ex: too many open files), so, the s pointer is nil */
	if s != nil {
		s.Stop()
	}
	return r, err
}

// Here is the gRPC server that GRPCClient talks to.
type ExecutorServer struct {
	// This is the real implementation
	sxproto.ExecutorServer
	Impl   Executor
	broker *goplugin.GRPCBroker
}

// Execute is where the magic happens
func (m ExecutorServer) Execute(ctx context.Context, req *sxproto.ExecuteRequest) (*sxproto.ExecuteResponse, error) {
	conn, err := m.broker.Dial(req.StatusServer)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	a := &GRPCStatusHelperClient{sxproto.NewStatusHelperClient(conn)}
	return m.Impl.Execute(req, a)
}

// GRPCStatusHelperClient is an implementation of status updates over RPC.
type GRPCStatusHelperClient struct{ client sxproto.StatusHelperClient }

func (m *GRPCStatusHelperClient) Update(b []byte, c bool) (int64, error) {
	resp, err := m.client.Update(context.Background(), &sxproto.StatusUpdateRequest{
		Output: b,
		Error:  c,
	})
	if err != nil {
		return 0, err
	}
	return resp.R, err
}

// GRPCStatusHelperServer is the gRPC server that GRPCClient talks to.
type GRPCStatusHelperServer struct {
	// This is the real implementation
	sxproto.StatusHelperServer
	Impl StatusHelper
}

func (m *GRPCStatusHelperServer) Update(ctx context.Context, req *sxproto.StatusUpdateRequest) (resp *sxproto.StatusUpdateResponse, err error) {
	r, err := m.Impl.Update(req.Output, req.Error)
	if err != nil {
		return nil, err
	}
	return &sxproto.StatusUpdateResponse{R: r}, err
}
