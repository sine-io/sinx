package plugin

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// See serve.go for serving plugins

// PluginMap should be used by clients for the map of plugins.
var PluginMap = map[string]plugin.Plugin{
	"processor": &ProcessorPlugin{},
	"executor":  &ExecutorPlugin{},
}

var embedPlugins = []string{"shell", "http"}

type Plugins struct {
	Processors map[string]Processor
	Executors  map[string]Executor
	// LogLevel   string
	NodeName string
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
	processors, err := plugin.Discover("sinx-processor-*", pluginDir)
	if err != nil {
		return err
	}

	// Look in /etc/sinx/plugins (or the used config path)
	executors, err := plugin.Discover("sinx-executor-*", pluginDir)
	if err != nil {
		return err
	}

	// Next, look in the same directory as the SinX executable, usually
	// /usr/local/bin. If found, this replaces what we found in the config path.
	exePath, err := osext.Executable()
	if err != nil {
		zlog.Error().Err(err).Msg("Error loading exe directory")
	} else {
		p, err := plugin.Discover("sinx-processor-*", filepath.Dir(exePath))
		if err != nil {
			return err
		}
		processors = append(processors, p...)

		e, err := plugin.Discover("sinx-executor-*", filepath.Dir(exePath))
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
	var config plugin.ClientConfig

	config.Cmd = exec.Command(path, args...)
	config.HandshakeConfig = Handshake
	config.Managed = true
	config.Plugins = PluginMap
	config.SyncStdout = os.Stdout
	config.SyncStderr = os.Stderr
	config.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "plugins",
		Level: hclog.LevelFromString(viper.GetString("log-level")),
		// Output: zlog.Logger,
		Output: zlog.Logger.Hook(
			zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
				var jsonFields map[string]any

				err := json.Unmarshal([]byte(msg), &jsonFields)
				if err != nil {
					e.Err(err)
				} else {
					jsonFields["level"] = jsonFields["@level"]
					delete(jsonFields, "@level")
					delete(jsonFields, "@timestamp")

					// TODO: I don't know how to delete message in zlog.Logger, i will do it later.
					e.Fields(jsonFields)
				}
			}),
		),
		JSONFormat: true,
	})

	switch pluginType {
	case ProcessorPluginName:
		config.AllowedProtocols = []plugin.Protocol{plugin.ProtocolNetRPC}
	case ExecutorPluginName:
		config.AllowedProtocols = []plugin.Protocol{plugin.ProtocolGRPC}
	}

	client := plugin.NewClient(&config)

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
