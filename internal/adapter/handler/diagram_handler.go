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
	"go.mongodb.org/mongo-driver/mongo"
)

type DiagramHandler struct {
	diagramService *service.DiagramService
	validator      *validation.ValidationEngine
}

func NewDiagramHandler(
	diagramService *service.DiagramService,
	validator *validation.ValidationEngine,
) *DiagramHandler {
	return &DiagramHandler{
		diagramService: diagramService,
		validator:      validator,
	}
}

// CreateDiagram creates a new diagram in a project
func (h *DiagramHandler) CreateDiagram(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.CreateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewValidationErrorResponse(validationErrors)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Parse parent diagram ID if provided
	var parentDiagramID *primitive.ObjectID
	if req.ParentDiagramID != nil && *req.ParentDiagramID != "" {
		parentID, err := primitive.ObjectIDFromHex(*req.ParentDiagramID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
			return
		}
		parentDiagramID = &parentID
	}

	// Create diagram
	diagram, err := h.diagramService.CreateDiagram(
		c.Request.Context(),
		projectID,
		userID,
		req.DiagramName,
		req.Description,
		parentDiagramID,
		req.EncryptedData,
		req.EncryptedDataSignature,
	)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to create diagram")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrDiagramAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to create diagram")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("diagram_id", diagram.ID.Hex()).
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Diagram created")

	// TODO: Get timestamps from mgod
	response := dto.ToDiagramResponse(diagram)
	c.JSON(http.StatusCreated, dto.NewAPIResponse(response, nil))
}

// ListDiagrams gets all diagrams for a project with pagination
func (h *DiagramHandler) ListDiagrams(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Get pagination params
	var params dto.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params = dto.DefaultPaginationParams()
	}
	params.Validate()

	rootOnly := c.Query("root_only") == "true"

	diagrams, totalCount, err := h.diagramService.ListDiagrams(
		c.Request.Context(),
		projectID,
		userID,
		rootOnly,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrDiagramAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to list diagrams")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Convert to responses
	responses := make([]dto.DiagramResponse, 0, len(diagrams))
	for _, diagram := range diagrams {
		// TODO: Get actual timestamps from mgod
		responses = append(responses, dto.ToDiagramResponse(diagram))
	}

	paginationMeta := dto.NewPaginationMeta(params, totalCount)
	c.JSON(http.StatusOK, dto.NewAPIResponseWithPagination(responses, &paginationMeta))
}

// GetDiagram gets a specific diagram
func (h *DiagramHandler) GetDiagram(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	diagramIDStr := c.Param("diagram_id")
	diagramID, err := primitive.ObjectIDFromHex(diagramIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	diagram, err := h.diagramService.GetDiagram(c.Request.Context(), diagramID, userID)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrDiagramAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("diagram_id", diagramID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to get diagram")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// TODO: Get actual timestamps from mgod
	response := dto.ToDiagramResponse(diagram)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// UpdateDiagram updates an existing diagram
func (h *DiagramHandler) UpdateDiagram(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	diagramIDStr := c.Param("diagram_id")
	diagramID, err := primitive.ObjectIDFromHex(diagramIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.UpdateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewValidationErrorResponse(validationErrors)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Update diagram
	diagram, err := h.diagramService.UpdateDiagram(
		c.Request.Context(),
		diagramID,
		userID,
		req.DiagramName,
		req.Description,
		req.EncryptedData,
		req.EncryptedDataSignature,
	)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("diagram_id", diagramID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to update diagram")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrDiagramAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("diagram_id", diagramID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to update diagram")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("diagram_id", diagramID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Diagram updated")

	// TODO: Get actual timestamps from mgod
	response := dto.ToDiagramResponse(diagram)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// DeleteDiagram deletes a diagram
func (h *DiagramHandler) DeleteDiagram(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	diagramIDStr := c.Param("diagram_id")
	diagramID, err := primitive.ObjectIDFromHex(diagramIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	err = h.diagramService.DeleteDiagram(c.Request.Context(), diagramID, userID)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("diagram_id", diagramID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to delete diagram")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrDiagramAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeDiagramAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("diagram_id", diagramID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to delete diagram")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("diagram_id", diagramID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Diagram deleted")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Diagram deleted successfully",
	}, nil))
}
