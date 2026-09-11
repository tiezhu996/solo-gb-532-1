package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Details any
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

func NewError(status int, code, message string, cause error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Cause: cause}
}

func BadRequest(code, message string, cause error) *AppError {
	return NewError(http.StatusBadRequest, code, message, cause)
}

func Unauthorized(message string) *AppError {
	return NewError(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", message, nil)
}

func Forbidden(message string) *AppError {
	return NewError(http.StatusForbidden, "AUTH_FORBIDDEN", message, nil)
}

func NotFound(entity string, cause error) *AppError {
	return NewError(http.StatusNotFound, "RESOURCE_NOT_FOUND", entity+"不存在", cause)
}

func Conflict(code, message string, cause error) *AppError {
	return NewError(http.StatusConflict, code, message, cause)
}

func Unprocessable(code, message string, cause error) *AppError {
	return NewError(http.StatusUnprocessableEntity, code, message, cause)
}

func Internal(cause error) *AppError {
	return NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时无法完成请求", cause)
}

func RequestID(c *gin.Context) string { return c.GetString("request_id") }

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "request_id": RequestID(c)})
}

func Page(c *gin.Context, data any, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"meta":       gin.H{"page": page, "page_size": pageSize, "total": total},
		"request_id": RequestID(c),
	})
}

func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = Internal(err)
	}
	c.Error(appErr)
	c.AbortWithStatusJSON(appErr.Status, gin.H{
		"error":      gin.H{"code": appErr.Code, "message": appErr.Message, "details": appErr.Details},
		"request_id": RequestID(c),
	})
}

func BindError(c *gin.Context, err error) {
	appErr := BadRequest("VALIDATION_ERROR", "请求字段无效", err)
	appErr.Details = gin.H{"validation": err.Error()}
	WriteError(c, appErr)
}
