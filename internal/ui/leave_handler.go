package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sxagent "github.com/sine-io/sinx/internal/agent"
)

func (h *HTTPTransport) leaveHandler(c *gin.Context) {
	if err := sxagent.StopAgent(h.agent); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
	renderJSON(c, http.StatusOK, h.agent.Peers)
}
