package response

import (
	"net/http"

	"github.com/sine-io/sinx/pkg/errorx"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    errorx.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errorx.ErrSuccess,
		Message: errorx.GetErrorMessage(errorx.ErrSuccess),
		Data:    data,
	})
}

func Error(c *gin.Context, err *errorx.Error) {
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Data,
	})
}

func ErrorWithCode(c *gin.Context, code errorx.ErrorCode, data ...interface{}) {
	err := errorx.NewWithCode(code, data...)
	Error(c, err)
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errorx.ErrInternalServer,
		Message: err.Error(),
	})
}
