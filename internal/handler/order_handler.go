package handler

import (
	"math"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
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
		response.HandleValidationError(c, err)
		return
	}

	order, customer, err := h.orderService.CreateWebOrder(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.CreatedWithMsg(c, "Đặt hàng thành công", gin.H{
		"order":    order,
		"customer": customer,
	})
}

func (h *OrderHandler) CreatePosOrder(c *gin.Context) {
	var req dto.CreatePosOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	order, customer, err := h.orderService.CreatePosOrder(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.CreatedWithMsg(c, "Tạo đơn hàng POS thành công", gin.H{
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
			response.HandleError(c, err)
			return
		}
		if order == nil {
			response.FailWithStatus(c, http.StatusNotFound, apperror.CodeOrderNotFound, "Không tìm thấy đơn hàng với mã yêu cầu")
			return
		}
		response.Success(c, []any{order})
		return
	}

	if phone != "" {
		orders, err := h.orderRepo.GetByPhone(c.Request.Context(), phone)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		response.Success(c, orders)
		return
	}

	response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Vui lòng cung cấp số điện thoại hoặc mã đơn hàng để tra cứu")
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	if limit <= 0 {
		limit = 50
	}

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
		response.HandleError(c, err)
		return
	}

	page := int(offset/limit) + 1
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response.SuccessWithMeta(c, orders, &response.Meta{
		Page:       page,
		PageSize:   int(limit),
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã đơn hàng không đúng định dạng")
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	order, customer, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMsg(c, "Cập nhật trạng thái đơn hàng thành công", gin.H{
		"order":    order,
		"customer": customer,
	})
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã đơn hàng không đúng định dạng")
		return
	}

	order, err := h.orderRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if order == nil {
		response.FailWithStatus(c, http.StatusNotFound, apperror.CodeOrderNotFound, "Không tìm thấy thông tin đơn hàng")
		return
	}
	response.Success(c, order)
}
