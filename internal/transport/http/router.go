package httptransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/local/cry-084/internal/middleware"
	"go.uber.org/zap"
)

func NewRouter(handler *Handler, logger *zap.Logger, timeout time.Duration) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.RequestID(), middleware.Recovery(logger), middleware.AccessLog(logger), middleware.SecurityHeaders(), middleware.CORS([]string{"http://localhost:5173"}), middleware.RateLimit(20, 40), middleware.Timeout(timeout))
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	api := router.Group("/api/v1", middleware.DemoAuth())
	api.POST("/inspection-tasks/:id/claim", handler.ClaimTask)
	api.POST("/inspection-tasks/:id/begin", handler.BeginTask)
	api.POST("/inspection-tasks/:id/results", handler.SubmitResult)
	api.POST("/hazards/:id/assign", handler.AssignHazard)
	api.POST("/hazards/:id/rectify", handler.RectifyHazard)
	api.POST("/hazards/:id/review", handler.ReviewHazard)
	return router
}
