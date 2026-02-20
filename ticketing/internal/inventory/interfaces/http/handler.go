package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketing/internal/inventory/application"
	"ticketing/internal/inventory/domain"
	"ticketing/internal/inventory/interfaces/dto"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(r *gin.Engine) {
	r.POST("/inventory/try-hold", h.tryHold)
	r.POST("/inventory/release-hold", h.releaseHold)
	r.POST("/inventory/confirm-hold", h.confirmHold)
	r.GET("/inventory/availability", h.availability)
}

func (h *Handler) tryHold(c *gin.Context) {
	var req dto.TryHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	state, err := h.service.TryHold(c.Request.Context(), application.TryHoldInput{
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
		Qty:          req.Qty,
		Capacity:     req.Capacity,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrInvalidQuantity) || errors.Is(err, domain.ErrBackpressure) {
			status = http.StatusBadRequest
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, state)
}

func (h *Handler) releaseHold(c *gin.Context) {
	var req dto.ReleaseHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	state, err := h.service.ReleaseHold(c.Request.Context(), application.ReleaseInput{
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrHoldNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, state)
}

func (h *Handler) confirmHold(c *gin.Context) {
	var req dto.ConfirmHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	state, err := h.service.ConfirmHold(c.Request.Context(), application.ConfirmInput{
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrHoldNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, state)
}

func (h *Handler) availability(c *gin.Context) {
	key := c.Query("partition_key")
	if key == "" {
		writeError(c, http.StatusBadRequest, "partition_key is required")
		return
	}
	available, ok, err := h.service.GetAvailability(c.Request.Context(), key)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(c, http.StatusNotFound, "partition not found")
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{
		"partition_key": key,
		"available":     available,
	})
}

func writeJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func writeError(c *gin.Context, status int, message string) {
	writeJSON(c, status, map[string]string{"error": message})
}
