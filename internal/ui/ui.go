package ui

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	uiPathPrefix  = "ui/"
	apiPathPrefix = "v1"
)

//go:embed ui-dist
var uiDist embed.FS

// UI registers UI specific routes on the gin RouterGroup.
func (h *HTTPTransport) UI(r *gin.RouterGroup, aclEnabled bool) {
	// If we are visiting from a browser redirect to the dashboard
	r.GET("/", func(c *gin.Context) {
		switch c.NegotiateFormat(gin.MIMEHTML) {
		case gin.MIMEHTML:
			c.Redirect(http.StatusSeeOther, "/ui/")
		default:
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	ui := r.Group("/" + uiPathPrefix)

	assets, err := fs.Sub(uiDist, "ui-dist")
	if err != nil {
		h.logger.Fatal().Err(err)
	}

	fp, err := assets.Open("index.html")
	if err != nil {
		h.logger.Fatal().Err(err)
	}

	indexText, err := io.ReadAll(fp)
	if err != nil {
		h.logger.Fatal().Err(err)
	}

	templ, err := template.New("index.html").Parse(string(indexText))
	if err != nil {
		h.logger.Fatal().Err(err)
	}

	h.Engine.SetHTMLTemplate(templ)

	ui.GET("/*filepath", func(ctx *gin.Context) {
		p := ctx.Param("filepath")
		f := strings.TrimPrefix(p, "/")
		_, err := assets.Open(f)
		if err == nil && p != "/" && p != "/index.html" {
			ctx.FileFromFS(p, http.FS(assets))
		} else {
			jobs, err := h.agent.JobDB.GetJobs(nil)
			if err != nil {
				h.logger.Error().Err(err)
			}
			var (
				totalJobs                                   = len(jobs)
				successfulJobs, failedJobs, untriggeredJobs int
			)
			for _, j := range jobs {
				if j.Status == "success" {
					successfulJobs++
				} else if j.Status == "failed" {
					failedJobs++
				} else if j.Status == "" {
					untriggeredJobs++
				}
			}
			l, err := h.agent.LeaderMember()
			ln := "no leader"
			if err != nil {
				h.logger.Error().Err(err)
			} else {
				ln = l.Name
			}
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"SINX_API_URL":          fmt.Sprintf("../%s", apiPathPrefix),
				"SINX_LEADER":           ln,
				"SINX_TOTAL_JOBS":       totalJobs,
				"SINX_FAILED_JOBS":      failedJobs,
				"SINX_UNTRIGGERED_JOBS": untriggeredJobs,
				"SINX_SUCCESSFUL_JOBS":  successfulJobs,
				"SINX_ACL_ENABLED":      aclEnabled,
			})
		}
	})
}
