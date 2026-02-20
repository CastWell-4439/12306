package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketing/internal/order/application"
	"ticketing/internal/order/domain"
	"ticketing/internal/order/interfaces/dto"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(r *gin.Engine) {
	r.POST("/orders", h.createOrder)
	r.POST("/orders/reserve", h.reserveOrder)
	r.POST("/orders/cancel", h.cancelOrder)
	r.POST("/payments/callback", h.paymentCallback)
	r.GET("/orders/get", h.getOrder)
}

func (h *Handler) createOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	order, err := h.service.CreateOrder(c.Request.Context(), application.CreateOrderInput{
		IdempotencyKey: req.IdempotencyKey,
		AmountCents:    req.AmountCents,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidAmount) {
			status = http.StatusBadRequest
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, order)
}

func (h *Handler) reserveOrder(c *gin.Context) {
	var req dto.ReserveOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	order, err := h.service.ReserveOrder(c.Request.Context(), application.ReserveOrderInput{
		OrderID:      req.OrderID,
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
		Qty:          req.Qty,
		Capacity:     req.Capacity,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidStateTransfer) {
			status = http.StatusConflict
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, order)
}

func (h *Handler) cancelOrder(c *gin.Context) {
	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	order, err := h.service.CancelOrder(c.Request.Context(), application.CancelOrderInput{
		OrderID:      req.OrderID,
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidStateTransfer) {
			status = http.StatusConflict
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, order)
}

func (h *Handler) paymentCallback(c *gin.Context) {
	var req dto.PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json")
		return
	}
	order, err := h.service.PaymentCallback(c.Request.Context(), application.PaymentCallbackInput{
		OrderID:       req.OrderID,
		ProviderTxnID: req.ProviderTxnID,
		Status:        req.Status,
		PartitionKey:  req.PartitionKey,
		HoldID:        req.HoldID,
		Signature:     req.Signature,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, application.ErrInvalidSignature) {
			status = http.StatusUnauthorized
		}
		if errors.Is(err, application.ErrInvalidPaymentStatus) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, domain.ErrInvalidStateTransfer) {
			status = http.StatusConflict
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, order)
}

func (h *Handler) getOrder(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		writeError(c, http.StatusBadRequest, "order_id is required")
		return
	}
	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, order)
}

func writeJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func writeError(c *gin.Context, status int, message string) {
	writeJSON(c, status, map[string]any{"error": message})
}
