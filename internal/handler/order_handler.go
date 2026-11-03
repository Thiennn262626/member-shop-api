package handler

import (
	"member-shop-api/internal/dto"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderHandler struct {
	orderService services.OrderService
	orderRepo    repository.OrderRepository
}

func NewOrderHandler(orderService services.OrderService, orderRepo repository.OrderRepository) *OrderHandler {
	return &OrderHandler{orderService: orderService, orderRepo: orderRepo}
}

func (h *OrderHandler) CreateWebOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, customer, err := h.orderService.CreateWebOrder(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order":    order,
		"customer": customer,
	})
}

func (h *OrderHandler) CreatePosOrder(c *gin.Context) {
	var req dto.CreatePosOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, customer, err := h.orderService.CreatePosOrder(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order":    order,
		"customer": customer,
	})
}

func (h *OrderHandler) TrackOrders(c *gin.Context) {
	phone := c.Query("phone")
	code := c.Query("code")

	if code != "" {
		order, err := h.orderRepo.GetByCode(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if order == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy đơn hàng"})
			return
		}
		c.JSON(http.StatusOK, []any{order})
		return
	}

	if phone != "" {
		orders, err := h.orderRepo.GetByPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "cần truyền phone hoặc code để tra cứu"})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	filter := repository.OrderFilter{
		Status:        c.Query("status"),
		PaymentStatus: c.Query("paymentStatus"),
		Search:        c.Query("search"),
		CustomerPhone: c.Query("phone"),
		Limit:         limit,
		Offset:        offset,
	}

	orders, total, err := h.orderRepo.GetList(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã đơn hàng không đúng"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, customer, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":    order,
		"customer": customer,
	})
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã đơn hàng không đúng"})
		return
	}

	order, err := h.orderRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy đơn hàng"})
		return
	}
	c.JSON(http.StatusOK, order)
}
