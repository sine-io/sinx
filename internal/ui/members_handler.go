package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	sxproto "github.com/sine-io/sinx/types"
)

func (h *HTTPTransport) membersHandler(c *gin.Context) {
	mems := []*sxproto.Member{}
	for _, m := range h.agent.Serf.Members() {
		id, _ := uuid.GenerateUUID()
		mid := &sxproto.Member{
			Member:     m,
			Id:         id,
			StatusText: m.Status.String(),
		}
		mems = append(mems, mid)
	}
	c.Header("X-Total-Count", strconv.Itoa(len(mems)))
	renderJSON(c, http.StatusOK, mems)
}
