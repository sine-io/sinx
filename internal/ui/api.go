package ui

import (
	"net/http"

	"github.com/gin-contrib/expvar"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	pretty = "pretty"
)

// APIRoutes registers the api routes on the gin RouterGroup.
func (h *HTTPTransport) APIRoutes(r *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	r.GET("/debug/vars", expvar.Handler())

	h.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	if h.config.EnablePrometheus {
		// Prometheus metrics scrape endpoint
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	r.GET("/v1", h.indexHandler)

	v1 := r.Group("/v1")
	v1.Use(middleware...)
	v1.GET("/", h.indexHandler)
	v1.GET("/members", h.membersHandler)
	v1.GET("/leader", h.leaderHandler)
	v1.GET("/isleader", h.isLeaderHandler)
	v1.POST("/leave", h.leaveHandler)
	v1.POST("/restore", h.restoreHandler)

	v1.GET("/busy", h.busyHandler)

	v1.POST("/jobs", h.jobCreateOrUpdateHandler)
	v1.PATCH("/jobs", h.jobCreateOrUpdateHandler)
	// Place fallback routes last
	v1.GET("/jobs", h.jobsHandler)

	jobs := v1.Group("/jobs")
	jobs.DELETE("/:job", h.jobDeleteHandler)
	jobs.POST("/:job", h.jobRunHandler)
	jobs.POST("/:job/run", h.jobRunHandler)
	jobs.POST("/:job/toggle", h.jobToggleHandler)
	jobs.PUT("/:job", h.jobCreateOrUpdateHandler)

	// Place fallback routes last
	jobs.GET("/:job", h.jobGetHandler)
	jobs.GET("/:job/executions", h.executionsHandler)
	jobs.GET("/:job/executions/:execution", h.executionHandler)
}

func renderJSON(c *gin.Context, status int, v interface{}) {
	if _, ok := c.GetQuery(pretty); ok {
		c.IndentedJSON(status, v)
	} else {
		c.JSON(status, v)
	}
}
