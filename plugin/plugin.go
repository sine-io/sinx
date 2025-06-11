package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"

	"github.com/rs/zerolog"

	sxlog "github.com/sine-io/sinx/log"
)

// See serve.go for serving plugins

// PluginMap should be used by clients for the map of plugins.
var PluginMap = map[string]goplugin.Plugin{
	"processor": &ProcessorPlugin{},
	"executor":  &ExecutorPlugin{},
}

var embedPlugins = []string{"shell", "http"}

type Plugins struct {
	Processors map[string]Processor
	Executors  map[string]Executor

	logger zerolog.Logger
}

func NewPlugins() *Plugins {
	return &Plugins{
		Processors: make(map[string]Processor),
		Executors:  make(map[string]Executor),

		logger: zerolog.New(zerolog.NewConsoleWriter()), // default logger
	}
}

func (p *Plugins) WithLogger(logger *zerolog.Logger) *Plugins {
	p.logger = logger.Hook() // TODO: do nothing now, but we can add hooks later

	return p
}

// DiscoverPlugins located on disk
//
// We look in the following places for plugins:
//
// 1. SinX configuration path
// 2. Path where SinX is installed
//
// Whichever file is discoverd LAST wins.
func (p *Plugins) DiscoverPlugins() error {
	p.Processors = make(map[string]Processor)
	p.Executors = make(map[string]Executor)

	pluginDir := filepath.Join("/etc", "sinx", "plugins")
	if viper.ConfigFileUsed() != "" {
		pluginDir = filepath.Join(filepath.Dir(viper.ConfigFileUsed()), "plugins")
	}

	// Look in /etc/sinx/plugins (or the used config path)
	processors, err := goplugin.Discover("sinx-processor-*", pluginDir)
	if err != nil {
		return err
	}

	// Look in /etc/sinx/plugins (or the used config path)
	executors, err := goplugin.Discover("sinx-executor-*", pluginDir)
	if err != nil {
		return err
	}

	// Next, look in the same directory as the SinX executable, usually
	// /usr/local/bin. If found, this replaces what we found in the config path.
	exePath, err := osext.Executable()
	if err != nil {
		p.logger.Error().Err(err).Msg("Error loading exe directory")
	} else {
		p, err := goplugin.Discover("sinx-processor-*", filepath.Dir(exePath))
		if err != nil {
			return err
		}
		processors = append(processors, p...)

		e, err := goplugin.Discover("sinx-executor-*", filepath.Dir(exePath))
		if err != nil {
			return err
		}
		executors = append(executors, e...)
	}

	for _, file := range processors {
		pluginName, ok := getPluginName(file)
		if !ok {
			continue
		}

		raw, err := p.pluginFactory(file, []string{}, ProcessorPluginName)
		if err != nil {
			return err
		}
		p.Processors[pluginName] = raw.(Processor)
	}

	for _, file := range executors {
		pluginName, ok := getPluginName(file)
		if !ok {
			continue
		}

		raw, err := p.pluginFactory(file, []string{}, ExecutorPluginName)
		if err != nil {
			return err
		}
		p.Executors[pluginName] = raw.(Executor)
	}

	// Load the embeded plugins
	for _, pluginName := range embedPlugins {
		raw, err := p.pluginFactory(exePath, []string{pluginName}, ExecutorPluginName)
		if err != nil {
			return err
		}
		p.Executors[pluginName] = raw.(Executor)
	}

	return nil
}

func getPluginName(file string) (string, bool) {
	// Look for foo-bar-baz. The plugin name is "baz"
	// sine, 2025.5.30
	// we should use filepath.Base instead of path.Base.
	// because path.Base will return '/', but filepath.Base will return '/' or '\\'.
	base := filepath.Base(file)
	parts := strings.SplitN(base, "-", 3)
	if len(parts) != 3 {
		return "", false
	}

	// This cleans off the .exe for windows plugins
	name := strings.TrimSuffix(parts[2], ".exe")
	return name, true
}

func (p *Plugins) pluginFactory(path string, args []string, pluginType string) (interface{}, error) {
	// Build the plugin client configuration and init the plugin
	var config = &goplugin.ClientConfig{
		Cmd:             exec.Command(path, args...),
		HandshakeConfig: Handshake,
		Managed:         true,
		Plugins:         PluginMap,
		SyncStdout:      os.Stdout,
		SyncStderr:      os.Stderr,
		Logger: sxlog.HclogWrapper(
			"plugins",
			viper.GetString("log-level"),
			&p.logger,
		),
	}

	switch pluginType {
	case ProcessorPluginName:
		config.AllowedProtocols = []goplugin.Protocol{goplugin.ProtocolNetRPC}
	case ExecutorPluginName:
		config.AllowedProtocols = []goplugin.Protocol{goplugin.ProtocolGRPC}
	}

	client := goplugin.NewClient(config)

	// Request the RPC client so we can get the provider
	// so we can build the actual RPC-implemented provider.
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginType)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
