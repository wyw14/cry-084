package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			var raw [8]byte
			_, _ = rand.Read(raw[:])
			id = hex.EncodeToString(raw[:])
		}
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "request_id", id))
		c.Next()
	}
}
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", zap.Any("panic", recovered), zap.String("request_id", c.GetString("request_id")))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "服务暂时不可用", "field_errors": gin.H{}, "request_id": c.GetString("request_id")})
	})
}
func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http request", zap.String("method", c.Request.Method), zap.String("path", c.FullPath()), zap.Int("status", c.Writer.Status()), zap.Duration("duration", time.Since(started)), zap.String("request_id", c.GetString("request_id")))
	}
}
