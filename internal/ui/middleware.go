package ui

import (
	"github.com/gin-contrib/cors"
	glogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// MetaMiddleware adds middleware to the gin Context.
func (h *HTTPTransport) MetaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Whom", h.config.NodeName)
		c.Next()
	}
}

func (h *HTTPTransport) CorsMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"*"}
	config.AllowHeaders = []string{"*"}
	config.ExposeHeaders = []string{"*"}

	return cors.New(config)
}

func (h *HTTPTransport) LoggerMiddleware() gin.HandlerFunc {
	return glogger.SetLogger( // TODO: we can set message, message default is "Request"
		glogger.WithLogger(func(c *gin.Context, _ zerolog.Logger) zerolog.Logger {
			return h.logger
		}),
	)
}
