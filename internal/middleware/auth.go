package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/local/cry-084/internal/domain/shared"
)

func DemoAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "请登录", "field_errors": gin.H{}, "request_id": c.GetString("request_id")})
			return
		}
		actor := shared.Actor{ID: shared.ID(c.GetHeader("X-Actor-ID")), MallID: shared.ID(c.GetHeader("X-Mall-ID")), TeamID: shared.ID(c.GetHeader("X-Team-ID")), Roles: strings.Split(c.GetHeader("X-Roles"), ","), Source: "http"}
		if actor.ID == "" || actor.MallID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "INVALID_IDENTITY", "message": "身份范围缺失", "field_errors": gin.H{}, "request_id": c.GetString("request_id")})
			return
		}
		c.Set("actor", actor)
		c.Next()
	}
}
