package ui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	status "google.golang.org/grpc/status"

	sxjob "github.com/sine-io/sinx/internal/job"
)

func (h *HTTPTransport) jobGetHandler(c *gin.Context) {
	jobName := c.Param("job")

	job, err := h.agent.Store.GetJob(jobName, nil)
	if err != nil {
		h.logger.Error().Err(err)
	}
	if job == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	renderJSON(c, http.StatusOK, job)
}

func (h *HTTPTransport) jobDeleteHandler(c *gin.Context) {
	jobName := c.Param("job")

	// Call gRPC DeleteJob
	job, err := h.agent.GRPCClient.DeleteJob(jobName)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}
	renderJSON(c, http.StatusOK, job)
}

func (h *HTTPTransport) jobRunHandler(c *gin.Context) {
	jobName := c.Param("job")

	// Call gRPC RunJob
	job, err := h.agent.GRPCClient.RunJob(jobName)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.Header("Location", c.Request.RequestURI)
	c.Status(http.StatusAccepted)
	renderJSON(c, http.StatusOK, job)
}

func (h *HTTPTransport) jobToggleHandler(c *gin.Context) {
	jobName := c.Param("job")

	job, err := h.agent.Store.GetJob(jobName, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// Toggle job status
	job.Disabled = !job.Disabled

	// Call gRPC SetJob
	if err := h.agent.GRPCClient.SetJob(job); err != nil {
		_ = c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	c.Header("Location", c.Request.RequestURI)
	renderJSON(c, http.StatusOK, job)
}

func (h *HTTPTransport) jobsHandler(c *gin.Context) {
	metadata := c.QueryMap("metadata")
	sort := c.DefaultQuery("_sort", "id")
	if sort == "id" {
		sort = "name"
	}
	order := c.DefaultQuery("_order", "ASC")
	q := c.Query("q")

	jobs, err := h.agent.Store.GetJobs(
		&JobOptions{
			Metadata: metadata,
			Sort:     sort,
			Order:    order,
			Query:    q,
			Status:   c.Query("status"),
			Disabled: c.Query("disabled"),
		},
	)
	if err != nil {
		h.logger.Error().Err(err).Msg("api: Unable to get jobs, store not reachable.")
		return
	}

	start, ok := c.GetQuery("_start")
	if !ok {
		start = "0"
	}
	s, _ := strconv.Atoi(start)

	end, ok := c.GetQuery("_end")
	e := 0
	if !ok {
		e = len(jobs)
	} else {
		e, _ = strconv.Atoi(end)
		if e > len(jobs) {
			e = len(jobs)
		}
	}

	c.Header("X-Total-Count", strconv.Itoa(len(jobs)))
	renderJSON(c, http.StatusOK, jobs[s:e])
}

func (h *HTTPTransport) jobCreateOrUpdateHandler(c *gin.Context) {
	// Init the Job object with defaults
	job := sxjob.Job{
		Concurrency: sxjob.ConcurrencyAllow,
	}

	// Check if the job is already in the context set by the middleware
	val, exists := c.Get("job")
	if exists {
		job = val.(sxjob.Job)
	} else {
		// Parse values from JSON
		if err := c.BindJSON(&job); err != nil {
			_, _ = c.Writer.WriteString(fmt.Sprintf("Unable to parse payload: %s.", err))
			h.logger.Error().Err(err)
			return
		}
		// Get the owner from the context, if it exists
		// this is coming from the ACL middleware
		accessor := c.GetString("accessor")
		if accessor != "" {
			job.Owner = accessor
		}
	}

	// Validate job
	if err := job.Validate(); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		_, _ = c.Writer.WriteString(fmt.Sprintf("Job contains invalid value: %s.", err))
		return
	}

	// Call gRPC SetJob
	if err := h.agent.GRPCClient.SetJob(&job); err != nil {
		s := status.Convert(err)
		if s.Message() == sxjob.ErrParentJobNotFound.Error() {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		_, _ = c.Writer.WriteString(s.Message())
		return
	}

	// Immediately run the job if so requested
	if _, exists := c.GetQuery("runoncreate"); exists {
		go func() {
			if _, err := h.agent.GRPCClient.RunJob(job.Name); err != nil {
				h.logger.Error().Err(err).Msg("api: Unable to run job.")
			}
		}()
	}

	c.Header("Location", fmt.Sprintf("%s/%s", c.Request.RequestURI, job.Name))
	renderJSON(c, http.StatusCreated, &job)
}
