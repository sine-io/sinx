package main

import (
	"os"
	"path/filepath"
	"strconv"

	zlog "github.com/rs/zerolog/log"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

const defaultLogDir = "/var/log/dkron"

// FilesOutput plugin that saves each execution log
// in it's own file in the file system.
type FilesOutput struct {
	forward bool
	logDir  string
}

// Process method writes the execution output to a file
func (l *FilesOutput) Process(args *sxplugin.ProcessorArgs) sxproto.Execution {
	l.parseConfig(args.Config)

	out := args.Execution.Output

	// Ensure our path exists
	if err := os.MkdirAll(l.logDir, 0o755); err != nil && !os.IsExist(err) {
		zlog.Error().Msgf("logDir path not accessible: %v", err)
	}
	// logFilepath := fmt.Sprintf("%s/%s.log", l.logDir, args.Execution.Key())
	// sine. 2025.5.30
	// Use filepath.Join to ensure correct path separators
	// This is important for cross-platform compatibility
	// and to avoid issues with different OS path formats.
	logFilepath := filepath.Join(l.logDir, args.Execution.Key()+".log")

	zlog.Info().Str("file", logFilepath).Msg("files: Writing file")
	if err := os.WriteFile(logFilepath, out, 0644); err != nil {
		zlog.Error().Err(err).Msg("Error writing log file")
	}

	if !l.forward {
		args.Execution.Output = []byte(logFilepath)
	}

	return args.Execution
}

func (l *FilesOutput) parseConfig(config sxplugin.Config) {
	forward, err := strconv.ParseBool(config["forward"])
	if err != nil {
		l.forward = false
		zlog.Warn().Str("param", "forward").Msg("Incorrect format or param not found.")
	} else {
		l.forward = forward
		zlog.Info().Msgf("Forwarding set to: %t", forward)
	}

	logDir := config["log_dir"]
	if logDir != "" {
		l.logDir = logDir
		zlog.Info().Msgf("Log dir set to: %s", logDir)
	} else {
		l.logDir = defaultLogDir
		zlog.Warn().Str("param", "log_dir").Msg("Incorrect format or param not found.")
		if _, err := os.Stat(defaultLogDir); os.IsNotExist(err) {
			os.MkdirAll(defaultLogDir, os.ModePerm)
		}
	}
}
