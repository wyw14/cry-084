package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/local/cry-084/internal/middleware"
	"github.com/gin-gonic/gin"
)

func perform(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestSecurityMiddlewareAddsStableHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := perform(router, http.MethodGet, "/healthz")
	if recorder.Code != http.StatusOK || recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("unexpected response: %d %q", recorder.Code, recorder.Header().Get("X-Content-Type-Options"))
	}
}
