package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}
func CORS(origins []string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, v := range origins {
		allowed[v] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type bucket struct {
	tokens  float64
	updated time.Time
}

func RateLimit(rate, burst float64) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]bucket{}
	return func(c *gin.Context) {
		key := strings.Split(c.Request.RemoteAddr, ":")[0]
		now := time.Now()
		mu.Lock()
		b := buckets[key]
		if b.updated.IsZero() {
			b = bucket{tokens: burst, updated: now}
		}
		b.tokens = min(burst, b.tokens+now.Sub(b.updated).Seconds()*rate)
		b.updated = now
		if b.tokens < 1 {
			buckets[key] = b
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁", "field_errors": gin.H{}, "request_id": c.GetString("request_id")})
			return
		}
		b.tokens--
		buckets[key] = b
		mu.Unlock()
		c.Next()
	}
}
