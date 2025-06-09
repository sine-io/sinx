package ui

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *HTTPTransport) busyHandler(c *gin.Context) {
	executions := []*Execution{}

	exs, err := h.agent.GetActiveExecutions()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, e := range exs {
		executions = append(executions, NewExecutionFromProto(e))
	}

	sort.SliceStable(executions, func(i, j int) bool {
		return executions[i].StartedAt.Before(executions[j].StartedAt)
	})

	c.Header("X-Total-Count", strconv.Itoa(len(executions)))
	renderJSON(c, http.StatusOK, executions)
}
