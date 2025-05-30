package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/distribworks/dkron/v4/plugin"
	"github.com/distribworks/dkron/v4/types"
	log "github.com/sirupsen/logrus"
)

const defaultLogDir = "/var/log/dkron"

// FilesOutput plugin that saves each execution log
// in it's own file in the file system.
type FilesOutput struct {
	forward bool
	logDir  string
}

// Process method writes the execution output to a file
func (l *FilesOutput) Process(args *plugin.ProcessorArgs) types.Execution {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	l.parseConfig(args.Config)

	out := args.Execution.Output

	// Ensure our path exists
	if err := os.MkdirAll(l.logDir, 0o755); err != nil && !os.IsExist(err) {
		log.Errorf("logDir path not accessible: %v", err)
	}
	// logFilepath := fmt.Sprintf("%s/%s.log", l.logDir, args.Execution.Key())
	// sine. 2025.5.30
	// Use filepath.Join to ensure correct path separators
	// This is important for cross-platform compatibility
	// and to avoid issues with different OS path formats.
	logFilepath := filepath.Join(l.logDir, args.Execution.Key()+".log")

	log.WithField("file", logFilepath).Info("files: Writing file")
	if err := os.WriteFile(logFilepath, out, 0644); err != nil {
		log.WithError(err).Error("Error writing log file")
	}

	if !l.forward {
		args.Execution.Output = []byte(logFilepath)
	}

	return args.Execution
}

func (l *FilesOutput) parseConfig(config plugin.Config) {
	forward, err := strconv.ParseBool(config["forward"])
	if err != nil {
		l.forward = false
		log.WithField("param", "forward").Warning("Incorrect format or param not found.")
	} else {
		l.forward = forward
		log.Infof("Forwarding set to: %t", forward)
	}

	logDir := config["log_dir"]
	if logDir != "" {
		l.logDir = logDir
		log.Infof("Log dir set to: %s", logDir)
	} else {
		l.logDir = defaultLogDir
		log.WithField("param", "log_dir").Warning("Incorrect format or param not found.")
		if _, err := os.Stat(defaultLogDir); os.IsNotExist(err) {
			os.MkdirAll(defaultLogDir, os.ModePerm)
		}
	}
}
