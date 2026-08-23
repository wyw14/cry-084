package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/local/cry-084/internal/domain/shared"
)

type errorBody struct {
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"field_errors"`
	RequestID   string            `json:"request_id"`
}

func writeError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(err, shared.ErrValidation):
		status, code, message = http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, shared.ErrForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "无权操作该资源"
	case errors.Is(err, shared.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "资源不存在"
	case errors.Is(err, shared.ErrConflict):
		status, code, message = http.StatusConflict, "VERSION_CONFLICT", "数据已变化，请刷新后重试"
	case errors.Is(err, shared.ErrInvalidState):
		status, code, message = http.StatusConflict, "INVALID_STATE", err.Error()
	}
	c.AbortWithStatusJSON(status, errorBody{Code: code, Message: message, FieldErrors: map[string]string{}, RequestID: c.GetString("request_id")})
}
