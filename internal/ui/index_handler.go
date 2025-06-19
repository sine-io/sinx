package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sxconfig "github.com/sine-io/sinx/internal/config"
)

func (h *HTTPTransport) indexHandler(c *gin.Context) {
	local := h.agent.Serf().LocalMember()

	stats := map[string]map[string]string{
		"agent": {
			"name":    local.Name,
			"version": sxconfig.Version,
		},
		"serf": h.agent.Serf().Stats(),
		"tags": local.Tags,
	}

	renderJSON(c, http.StatusOK, stats)
}
