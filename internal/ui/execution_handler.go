package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/buntdb"
)

type apiExecution struct {
	*Execution
	OutputTruncated bool `json:"output_truncated"`
}

func (h *HTTPTransport) executionsHandler(c *gin.Context) {
	jobName := c.Param("job")

	sort := c.DefaultQuery("_sort", "")
	if sort == "id" {
		sort = "started_at"
	}
	order := c.DefaultQuery("_order", "DESC")
	outputSizeLimit, err := strconv.Atoi(c.DefaultQuery("output_size_limit", ""))
	if err != nil {
		outputSizeLimit = -1
	}

	job, err := h.agent.Store.GetJob(jobName, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	executions, err := h.agent.Store.GetExecutions(job.Name,
		&ExecutionOptions{
			Sort:     sort,
			Order:    order,
			Timezone: job.GetTimeLocation(),
		},
	)
	if err == buntdb.ErrNotFound {
		executions = make([]*Execution, 0)
	} else if err != nil {
		h.logger.Error().Err(err)
		return
	}

	apiExecutions := make([]*apiExecution, len(executions))
	for j, execution := range executions {
		apiExecutions[j] = &apiExecution{execution, false}
		if outputSizeLimit > -1 {
			// truncate execution output
			size := len(execution.Output)
			if size > outputSizeLimit {
				apiExecutions[j].Output = apiExecutions[j].Output[size-outputSizeLimit:]
				apiExecutions[j].OutputTruncated = true
			}
		}
	}

	c.Header("X-Total-Count", strconv.Itoa(len(executions)))
	renderJSON(c, http.StatusOK, apiExecutions)
}

func (h *HTTPTransport) executionHandler(c *gin.Context) {
	jobName := c.Param("job")
	executionName := c.Param("execution")

	job, err := h.agent.Store.GetJob(jobName, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	executions, err := h.agent.Store.GetExecutions(job.Name,
		&ExecutionOptions{
			Sort:     "",
			Order:    "",
			Timezone: job.GetTimeLocation(),
		},
	)

	if err != nil {
		h.logger.Error().Err(err)
		return
	}

	for _, execution := range executions {
		if execution.Id == executionName {
			renderJSON(c, http.StatusOK, execution)
			return
		}
	}
}
