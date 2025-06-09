package ui

import (
	"io"

	"github.com/gin-contrib/cors"
	glogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	sxagent "github.com/sine-io/sinx/internal/agent"
)

type Transport interface {
	ServeHTTP()
}

type HTTPTransport struct {
	Engine *gin.Engine
	agent  *sxagent.Agent
}

func NewHTTPTransport(agent *sxagent.Agent) *HTTPTransport {
	return &HTTPTransport{
		Engine: gin.Default(),
		agent:  agent,
	}
}

func (h *HTTPTransport) ServeHTTP() {
	if h.agent.Config.LogLevel == "debug" {
		gin.DefaultWriter = h.agent.Logger.Hook(
			zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
				e.Str("level", "debug") // Add log level to each event
			}),
		)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.DefaultWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	}

	rootPath := h.Engine.Group("/")

	// Set up CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"*"}
	config.AllowHeaders = []string{"*"}
	config.ExposeHeaders = []string{"*"}
	rootPath.Use(cors.New(config))

	// Set up metadata handler
	rootPath.Use(h.MetaMiddleware())

	// Set up logging middleware
	rootPath.Use(glogger.SetLogger( // TODO: we can set message, message default is "Request"
		glogger.WithLogger(func(c *gin.Context, _ zerolog.Logger) zerolog.Logger {
			return h.agent.Logger
		})))

	h.APIRoutes(rootPath)
	if h.agent.Config.UI {
		h.UI(rootPath, false)
	}

	h.agent.Logger.Info().Str("address", h.agent.Config.HTTPAddr).Msg("api: Running HTTP server")

	go func() {
		if err := h.Engine.Run(h.agent.Config.HTTPAddr); err != nil {
			h.agent.Logger.Error().Err(err).Msg("api: Error starting HTTP server")
		}
	}()
}
