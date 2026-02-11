package handler

import (
	"errors"
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/gin-gonic/gin"
)

type BreadcrumbHandler struct {
	service *service.BreadcrumbService
}

func NewBreadcrumbHandler(service *service.BreadcrumbService) *BreadcrumbHandler {
	return &BreadcrumbHandler{service: service}
}

// GetBreadcrumbs godoc
// @Summary Get breadcrumbs for a resource
// @Tags projects
// @Produce json
// @Param project_id path string true "Project ID"
// @Param type query string true "Resource Type (project, note, diagram, node, vault)"
// @Param id query string false "Resource ID"
// @Success 200 {object} dto.APIResponse[dto.BreadcrumbResponse]
// @Router /api/v1/projects/{project_id}/breadcrumbs [get]
func (h *BreadcrumbHandler) GetBreadcrumbs(c *gin.Context) {
	projectID := c.Param("project_id")
	resourceType := c.Query("type")
	resourceID := c.Query("id")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Project ID is required")))
		return
	}

	if resourceType == "" {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Resource type is required")))
		return
	}

	breadcrumbs, err := h.service.GetBreadcrumbs(c.Request.Context(), projectID, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid ID format")))
		} else if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectNotFound)))
		} else if errors.Is(err, service.ErrResourceNotFound) {
			// Return the wrapped error message for debugging
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodePageNotFound, err.Error())))
		} else if errors.Is(err, service.ErrInvalidResourceType) {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid resource type")))
		} else {
			logger.Error().Err(err).Msg("Failed to get breadcrumbs")
			c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInternalError)))
		}
		return
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(breadcrumbs, nil))
}
