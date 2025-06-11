package ui

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	sxagent "github.com/sine-io/sinx/internal/agent"
	sxconfig "github.com/sine-io/sinx/internal/config"
)

type HTTPTransport struct {
	Engine *gin.Engine
	config *sxconfig.Config
	agent  *sxagent.Agent
	logger zerolog.Logger
}

func NewHTTPTransport(agent *sxagent.Agent) *HTTPTransport {
	return &HTTPTransport{
		Engine: gin.Default(),
		agent:  agent,

		logger: zerolog.New(zerolog.NewConsoleWriter()),
	}
}

func (h *HTTPTransport) WithLogger(logger *zerolog.Logger) *HTTPTransport {
	h.logger = logger.Hook()

	return h
}

func (h *HTTPTransport) ServeHTTP() {
	if viper.GetString("log-level") == "debug" {
		// gin.DefaultWriter = h.logger.Hook(
		// 	zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		// 		e.Str("level", "debug") // Add log level to each event
		// 	}),
		// )
		gin.DefaultWriter = os.Stderr
		gin.SetMode(gin.DebugMode)
	} else {
		gin.DefaultWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	}

	rootPath := h.Engine.Group("/")

	// Set up CORS
	rootPath.Use(h.CorsMiddleware())

	// Set up metadata handler
	rootPath.Use(h.MetaMiddleware())

	// Set up logging middleware
	rootPath.Use(h.LoggerMiddleware())

	h.APIRoutes(rootPath)

	if h.config.UI {
		h.UI(rootPath, false)
	}

	h.logger.Info().Str("address", h.config.HTTPAddr).Msg("api: Running HTTP server")

	go func() {
		if err := h.Engine.Run(h.config.HTTPAddr); err != nil {
			h.logger.Error().Err(err).Msg("api: Error starting HTTP server")
		}
	}()
}
