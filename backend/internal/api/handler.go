package api

import (
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sarosobserv/sarosobserv/backend/internal/telemetry"
)

// Handler handles API requests.
type Handler struct {
	storage telemetry.Storage
}

// NewHandler creates a new API handler.
func NewHandler(storage telemetry.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// RegisterRoutes registers API routes.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/services", h.ListServices)
		api.GET("/traces", h.ListTraces)
		api.GET("/traces/:traceId", h.GetTrace)
		api.GET("/traces/:traceId/logs", h.ListTraceLogs)
		api.GET("/overview/requests", h.RequestOverview)
	}
}

// Health returns health status.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "sarosobserv",
	})
}

// ListServices returns discovered services.
func (h *Handler) ListServices(c *gin.Context) {
	services, err := h.storage.ListServices(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
	})
}

// ListTraces returns a list of traces.
func (h *Handler) ListTraces(c *gin.Context) {
	limit := 50

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}

		limit = parsed
	}

	traces, err := h.storage.ListTraces(
		c.Request.Context(),
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"traces": traces,
	})
}

// GetTrace returns a specific trace.
func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceId")

	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "trace ID is required",
		})
		return
	}

	spans, err := h.storage.GetTrace(
		c.Request.Context(),
		traceID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(spans) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "trace not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trace_id": traceID,
		"spans":    spans,
	})
}

func (h *Handler) ListTraceLogs(c *gin.Context) {
	traceID := strings.ToLower(c.Param("traceId"))
	if _, err := hex.DecodeString(traceID); len(traceID) != 32 || err != nil || traceID == strings.Repeat("0", 32) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a valid trace ID is required"})
		return
	}
	records, err := h.storage.ListTraceLogs(c.Request.Context(), traceID)
	if err != nil {
		log.Printf("List trace logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load trace logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": records})
}

func (h *Handler) RequestOverview(c *gin.Context) {
	result, err := h.storage.RequestOverview(c.Request.Context(), time.Now())
	if err != nil {
		log.Printf("Load request overview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load request overview"})
		return
	}
	c.JSON(http.StatusOK, result)
}
