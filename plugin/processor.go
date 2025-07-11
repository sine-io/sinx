package plugin

import (
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"

	sxproto "github.com/sine-io/sinx/types"
)

// Execution processor.

// Processor is an interface that wraps the Process method.
// Plugins must implement this interface.
type Processor interface {
	// Main plugin method, will be called when an execution is done.
	Process(args *ProcessorArgs) sxproto.Execution
}

// ProcessorPlugin RPC implementation
type ProcessorPlugin struct {
	Processor Processor
}

// Server implements the RPC server
func (p *ProcessorPlugin) Server(b *goplugin.MuxBroker) (interface{}, error) {
	return &ProcessorServer{Broker: b, Processor: p.Processor}, nil
}

// Client implements the RPC client
func (p *ProcessorPlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProcessorClient{Broker: b, Client: c}, nil
}

// ProcessorArgs holds the Execution and PluginConfig for a Processor.
type ProcessorArgs struct {
	// The execution to pass to the processor
	Execution sxproto.Execution
	// The configuration for this plugin call
	Config Config
}

// Config holds a map of the plugin configuration data structure.
type Config map[string]string

// ProcessorClient is an implementation that talks over RPC
type ProcessorClient struct {
	Broker *goplugin.MuxBroker
	Client *rpc.Client
}

// Process method that actually call the plugin Process method.
func (e *ProcessorClient) Process(args *ProcessorArgs) sxproto.Execution {
	var resp sxproto.Execution
	err := e.Client.Call("Plugin.Process", args, &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

// ProcessorServer is the RPC server that client talks to, conforming to
// the requirements of net/rpc
type ProcessorServer struct {
	// This is the real implementation
	Broker    *goplugin.MuxBroker
	Processor Processor
}

// Process will call the actual Process method of the plugin
func (e *ProcessorServer) Process(args *ProcessorArgs, resp *sxproto.Execution) error {
	*resp = e.Processor.Process(args)
	return nil
}
