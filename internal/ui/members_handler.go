package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"

	sxproto "github.com/sine-io/sinx/types"
)

func (h *HTTPTransport) membersHandler(c *gin.Context) {
	var members []*sxproto.Member
	for _, m := range h.agent.Serf().Members() {
		id, _ := uuid.GenerateUUID()
		mid := &sxproto.Member{
			Member:     m,
			Id:         id,
			StatusText: m.Status.String(),
		}
		members = append(members, mid)
	}
	c.Header("X-Total-Count", strconv.Itoa(len(members)))
	renderJSON(c, http.StatusOK, members)
}
