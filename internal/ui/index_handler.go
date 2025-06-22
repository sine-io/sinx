package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sxcfg "github.com/sine-io/sinx/internal/config"
)

func (h *HTTPTransport) indexHandler(c *gin.Context) {
	local := h.agent.Serf().LocalMember()

	stats := map[string]map[string]string{
		"agent": {
			"name":    local.Name,
			"version": sxcfg.Version,
		},
		"serf": h.agent.Serf().Stats(),
		"tags": local.Tags,
	}

	renderJSON(c, http.StatusOK, stats)
}
