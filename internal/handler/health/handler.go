package health

import (
	"context"
	"net/http"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

type Checker interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	router  gin.IRouter
	checker Checker
	timeout time.Duration
}

func NewHandler(router gin.IRouter, checker Checker, timeout time.Duration) *Handler {
	return &Handler{
		router:  router,
		checker: checker,
		timeout: timeout,
	}
}

func (h *Handler) RouteList() {
	h.router.GET("/check-health", h.Check)
}

func (h *Handler) Check(c *gin.Context) {
	if h.checker == nil {
		httpresponse.JSONError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.checker.PingContext(ctx); err != nil {
		httpresponse.JSONError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	httpresponse.JSON(c, http.StatusOK, gin.H{
		"message": "service is healthy",
	})
}
