package ui

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"

	sxexec "github.com/sine-io/sinx/internal/execution"
)

func (h *HTTPTransport) busyHandler(c *gin.Context) {
	var executions []*sxexec.Execution

	exs, err := h.agent.GetActiveExecutions()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, e := range exs {
		executions = append(executions, sxexec.NewExecutionFromProto(e))
	}

	sort.SliceStable(executions, func(i, j int) bool {
		return executions[i].StartedAt.Before(executions[j].StartedAt)
	})

	c.Header("X-Total-Count", strconv.Itoa(len(executions)))
	renderJSON(c, http.StatusOK, executions)
}
