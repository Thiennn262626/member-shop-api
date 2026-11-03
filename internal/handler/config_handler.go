package handler

import (
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ConfigHandler struct {
	repo repository.ConfigRepository
}

func NewConfigHandler(repo repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{repo: repo}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	cfg, err := h.repo.GetConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, cfg)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	delete(updates, "_id")
	delete(updates, "id")

	updated, err := h.repo.UpdateConfig(c.Request.Context(), updates)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMsg(c, "Cập nhật cấu hình cửa hàng thành công", updated)
}
