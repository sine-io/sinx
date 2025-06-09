package ui

import "github.com/gin-gonic/gin"

// MetaMiddleware adds middleware to the gin Context.
func (h *HTTPTransport) MetaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Whom", h.agent.Config.NodeName)
		c.Next()
	}
}
