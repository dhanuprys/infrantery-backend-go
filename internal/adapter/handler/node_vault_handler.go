package handler

import (
	"errors"
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NodeVaultHandler struct {
	service   *service.NodeVaultService
	validator *validation.ValidationEngine
}

func NewNodeVaultHandler(service *service.NodeVaultService, validator *validation.ValidationEngine) *NodeVaultHandler {
	return &NodeVaultHandler{
		service:   service,
		validator: validator,
	}
}

func (h *NodeVaultHandler) CreateVaultItem(c *gin.Context) {
	// Parse params
	nodeID := c.Param("node_id")
	projectIDStr := c.Param("project_id")
	if nodeID == "" || projectIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Node ID and Project ID are required")))
		return
	}

	// logger.Info().Str("node_id", nodeID).Str("project_id", projectIDStr).Msg("CreateVaultItem called")

	var req dto.CreateNodeVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to bind JSON in CreateVaultItem")
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, err.Error())))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
	projectID, _ := primitive.ObjectIDFromHex(projectIDStr)

	vaultItem, err := h.service.CreateVaultItem(c.Request.Context(), nodeID, projectID, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrVaultAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultAccessDenied)))
			return
		}
		logger.Error().Err(err).Msg("Failed to create vault item")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToNodeVaultResponse(vaultItem)
	c.JSON(http.StatusCreated, dto.NewAPIResponse(response, nil))
}

func (h *NodeVaultHandler) ListVaultItems(c *gin.Context) {
	// Parse params
	nodeID := c.Param("node_id")
	projectIDStr := c.Param("project_id")
	if nodeID == "" || projectIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Node ID and Project ID are required")))
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
	projectID, _ := primitive.ObjectIDFromHex(projectIDStr)

	items, err := h.service.ListVaultItems(c.Request.Context(), nodeID, projectID, userID)
	if err != nil {
		if errors.Is(err, service.ErrVaultAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultAccessDenied)))
			return
		}
		logger.Error().Err(err).Msg("Failed to list vault items")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	responses := make([]dto.NodeVaultResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, dto.ToNodeVaultResponse(item))
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(responses, nil))
}

func (h *NodeVaultHandler) GetVaultItem(c *gin.Context) {
	vaultID := c.Param("vault_id")

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	item, err := h.service.GetVaultItem(c.Request.Context(), vaultID, userID)
	if err != nil {
		if errors.Is(err, service.ErrVaultAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrVaultItemNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultItemNotFound)))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToNodeVaultResponse(item)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

func (h *NodeVaultHandler) UpdateVaultItem(c *gin.Context) {
	vaultID := c.Param("vault_id")

	var req dto.UpdateNodeVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	item, err := h.service.UpdateVaultItem(c.Request.Context(), vaultID, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrVaultAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrVaultItemNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultItemNotFound)))
			return
		}
		logger.Error().Err(err).Msg("Failed to update vault item")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToNodeVaultResponse(item)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

func (h *NodeVaultHandler) DeleteVaultItem(c *gin.Context) {
	vaultID := c.Param("vault_id")

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	err := h.service.DeleteVaultItem(c.Request.Context(), vaultID, userID)
	if err != nil {
		if errors.Is(err, service.ErrVaultAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrVaultItemNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeVaultItemNotFound)))
			return
		}
		logger.Error().Err(err).Msg("Failed to delete vault item")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse[any](nil, nil))
}
