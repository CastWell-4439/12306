package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketing/internal/query/application"
	"ticketing/internal/query/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/query/orders", h.getOrderView)
}

func (h *Handler) getOrderView(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		writeError(c, http.StatusBadRequest, "order_id is required")
		return
	}
	v, err := h.service.GetOrderView(c.Request.Context(), orderID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrOrderViewNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, http.StatusOK, v)
}

func writeJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func writeError(c *gin.Context, status int, message string) {
	writeJSON(c, status, map[string]string{"error": message})
}
