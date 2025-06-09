package ui

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	sxjob "github.com/sine-io/sinx/internal/job"
)

// Restore jobs from file.
// Overwrite job if the job is exist.
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
	var jobs []*sxjob.Job
	err = json.Unmarshal(data, &jobs)

	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jobTree, err := sxjob.GenerateJobTree(jobs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	result := sxjob.RecursiveSetJob(jobTree)
	resp, err := json.Marshal(result)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	renderJSON(c, http.StatusOK, string(resp))
}
