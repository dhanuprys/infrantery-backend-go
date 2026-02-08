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

type NodeHandler struct {
	nodeService *service.NodeService
	validator   *validation.ValidationEngine
}

func NewNodeHandler(nodeService *service.NodeService, validator *validation.ValidationEngine) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
		validator:   validator,
	}
}

// GetOrCreateNode gets a node or creates it if it doesn't exist
func (h *NodeHandler) GetOrCreateNode(c *gin.Context) {
	diagramIDStr := c.Param("diagram_id")
	diagramID, err := primitive.ObjectIDFromHex(diagramIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	nodeIDStr := c.Param("node_id")
	if nodeIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Node ID is required")))
		return
	}

	// Validate node ID format using DTO validation mostly for consistency but simpler to check here or in service
	// Service checks it too.

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	node, err := h.nodeService.GetOrCreateNode(c.Request.Context(), nodeIDStr, diagramID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNodeAccessDenied) || errors.Is(err, service.ErrInvalidNodeID) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrNodeNotFound) {
			// Should ideally not happen with GetOrCreate unless diagram not found or permission issue
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeNotFound)))
			return
		}
		// Check for diagram access denied specifically if coming from verifyDiagramPermission
		// But service returns generic errors often.
		// If diagram not found, service returns ErrCodeDiagramNotFound error (wrapped/new)

		logger.Error().Err(err).Str("node_id", nodeIDStr).Msg("Failed to get/create node")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToNodeResponse(node)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// UpdateNode updates a node
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	var req dto.UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	nodeIDStr := c.Param("node_id")

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	node, err := h.nodeService.UpdateNode(c.Request.Context(), nodeIDStr, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrNodeAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeNotFound)))
			return
		}
		logger.Error().Err(err).Str("node_id", nodeIDStr).Msg("Failed to update node")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToNodeResponse(node)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// DeleteNode deletes a node
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	nodeIDStr := c.Param("node_id")

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	err := h.nodeService.DeleteNode(c.Request.Context(), nodeIDStr, userID)
	if err != nil {
		if errors.Is(err, service.ErrNodeAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNodeNotFound)))
			return
		}
		logger.Error().Err(err).Str("node_id", nodeIDStr).Msg("Failed to delete node")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse[any](nil, nil))
}
