package ui

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	sxagent "github.com/sine-io/sinx/internal/agent"
)

// Restore jobs from file.
// Overwrite job if the job is existed.
func (h *HTTPTransport) restoreHandler(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var jobs []*sxagent.Job
	err = json.Unmarshal(data, &jobs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jobTree, err := sxagent.GenerateJobTree(jobs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result := sxagent.RecursiveSetJob(jobTree)
	resp, err := json.Marshal(result)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	renderJSON(c, http.StatusOK, string(resp))
}
