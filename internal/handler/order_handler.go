package handler

import (
	"math"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderHandler struct {
	orderService services.OrderService
	orderRepo    repository.OrderRepository
	cache        cache.RedisCacheService
}

func NewOrderHandler(orderService services.OrderService, orderRepo repository.OrderRepository, cacheSvc cache.RedisCacheService) *OrderHandler {
	return &OrderHandler{orderService: orderService, orderRepo: orderRepo, cache: cacheSvc}
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
		cacheKey := "shop:order:code:" + code
		order, err := cache.CacheAside(c.Request.Context(), h.cache, cacheKey, 1*time.Minute, func() (*model.Order, error) {
			return h.orderRepo.GetByCode(c.Request.Context(), code)
		})
		if err != nil {
			response.HandleError(c, err)
			return
		}
		if order == nil {
			response.FailWithStatus(c, http.StatusNotFound, apperror.CodeOrderNotFound, "Không tìm thấy đơn hàng với mã yêu cầu")
			return
		}

		// Bảo vệ dữ liệu cá nhân (SEC-04): Nếu không xác thực được số điện thoại đặt hàng, mask PII
		resOrder := *order
		if phone == "" || phone != order.CustomerPhone {
			resOrder.CustomerName = utils.MaskName(order.CustomerName)
			resOrder.CustomerPhone = utils.MaskPhone(order.CustomerPhone)
			resOrder.DeliveryAddress = utils.MaskAddress(order.DeliveryAddress)
		}

		response.Success(c, []any{resOrder})
		return
	}

	if phone != "" {
		orders, err := h.orderRepo.GetByPhone(c.Request.Context(), phone)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		if orders == nil {
			orders = []model.Order{}
		}

		// Bảo vệ dữ liệu cá nhân (SEC-04): Mask PII cho danh sách đơn theo phone
		maskedOrders := make([]model.Order, len(orders))
		for i, o := range orders {
			maskedO := o
			maskedO.CustomerName = utils.MaskName(o.CustomerName)
			maskedO.CustomerPhone = utils.MaskPhone(o.CustomerPhone)
			maskedO.DeliveryAddress = utils.MaskAddress(o.DeliveryAddress)
			maskedOrders[i] = maskedO
		}

		response.Success(c, maskedOrders)
		return
	}

	response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Vui lòng cung cấp số điện thoại hoặc mã đơn hàng để tra cứu")
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("pageSize", c.DefaultQuery("limit", "50")), 10, 64)
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	var offset int64
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, _ = strconv.ParseInt(offsetStr, 10, 64)
		if offset < 0 {
			offset = 0
		}
		page = (offset / limit) + 1
	} else {
		offset = (page - 1) * limit
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
	if orders == nil {
		orders = []model.Order{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 && total == 0 {
		totalPages = 1
	}

	response.SuccessWithMeta(c, orders, &response.Meta{
		Page:       int(page),
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
