package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HTTPTransport) leaderHandler(c *gin.Context) {
	member, err := h.agent.LeaderMember()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
	if member == nil {
		c.AbortWithStatus(http.StatusNotFound)
	}
	renderJSON(c, http.StatusOK, member)
}

func (h *HTTPTransport) isLeaderHandler(c *gin.Context) {
	leader := h.agent.IsLeader()
	if leader {
		renderJSON(c, http.StatusOK, "I am a leader")
	} else {
		renderJSON(c, http.StatusNotFound, "I am a follower")
	}
}
