package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HTTPTransport) leaveHandler(c *gin.Context) {
	if err := h.agent.StopAgent(); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
	renderJSON(c, http.StatusOK, h.agent.Peers())
}
