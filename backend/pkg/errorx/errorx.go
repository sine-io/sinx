package errorx

import (
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	// 通用错误码 10000-19999
	ErrSuccess        ErrorCode = 0
	ErrInternalServer ErrorCode = 10001
	ErrInvalidParam   ErrorCode = 10002
	ErrUnauthorized   ErrorCode = 10003
	ErrForbidden      ErrorCode = 10004
	ErrNotFound       ErrorCode = 10005
	ErrHasChildren    ErrorCode = 10006

	// 用户相关错误码 20000-29999
	ErrUserNotFound        ErrorCode = 20001
	ErrUserAlreadyExists   ErrorCode = 20002
	ErrUserInvalidPassword ErrorCode = 20003
	ErrUserInvalidToken    ErrorCode = 20004
	ErrUserTokenExpired    ErrorCode = 20005
)

type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func New(code ErrorCode, message string, data ...interface{}) *Error {
	err := &Error{
		Code:    code,
		Message: message,
	}
	if len(data) > 0 {
		err.Data = data[0]
	}
	return err
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrSuccess:
		return http.StatusOK
	case ErrInvalidParam:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrUserInvalidToken, ErrUserTokenExpired:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound, ErrUserNotFound:
		return http.StatusNotFound
	case ErrUserAlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// 预定义错误消息
var errorMessages = map[ErrorCode]string{
	ErrSuccess:             "success",
	ErrInternalServer:      "internal server error",
	ErrInvalidParam:        "invalid parameter",
	ErrUnauthorized:        "unauthorized",
	ErrForbidden:           "forbidden",
	ErrNotFound:            "not found",
	ErrHasChildren:         "resource has children",
	ErrUserNotFound:        "user not found",
	ErrUserAlreadyExists:   "user already exists",
	ErrUserInvalidPassword: "invalid password",
	ErrUserInvalidToken:    "invalid token",
	ErrUserTokenExpired:    "token expired",
}

func GetErrorMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "unknown error"
}

func NewWithCode(code ErrorCode, data ...interface{}) *Error {
	return New(code, GetErrorMessage(code), data...)
}
