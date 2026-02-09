package handler

import (
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
		switch err {
		case service.ErrInvalidID:
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid ID format")))
		case service.ErrProjectNotFound:
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectNotFound)))
		case service.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodePageNotFound, "Resource not found"))) // Reuse PageNotFound or create specific
		case service.ErrInvalidResourceType:
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid resource type")))
		default:
			logger.Error().Err(err).Msg("Failed to get breadcrumbs")
			c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInternalError)))
		}
		return
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(breadcrumbs, nil))
}
