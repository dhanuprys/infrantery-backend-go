package handler

import (
	"errors"
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	userRepo       port.UserRepository
	validator      *validation.ValidationEngine
}

func NewProjectHandler(
	projectService *service.ProjectService,
	userRepo port.UserRepository,
	validator *validation.ValidationEngine,
) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		userRepo:       userRepo,
		validator:      validator,
	}
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.CreateProjectRequest
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

	// Create project
	project, err := h.projectService.CreateProject(
		c.Request.Context(),
		userID,
		req.Name,
		req.Description,
		req.EncryptionSalt,
		req.EncryptedPrivateKey,
		req.EncryptionPublicKey,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to create project")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", project.ID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Project created")

	response := dto.ToProjectResponse(project)
	c.JSON(http.StatusCreated, dto.NewAPIResponse(response, nil))
}

// GetUserProjects gets all projects for the current user with pagination
func (h *ProjectHandler) GetUserProjects(c *gin.Context) {
	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Get pagination params
	var params dto.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params = dto.DefaultPaginationParams()
	}
	params.Validate()

	projects, totalCount, err := h.projectService.GetUserProjects(
		c.Request.Context(),
		userID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to get user projects")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Convert to responses
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, project := range projects {
		responses = append(responses, dto.ToProjectResponse(project))
	}

	paginationMeta := dto.NewPaginationMeta(params, totalCount)
	c.JSON(http.StatusOK, dto.NewAPIResponseWithPagination(responses, &paginationMeta))
}

// GetProjectDetails gets project details with user permissions
func (h *ProjectHandler) GetProjectDetails(c *gin.Context) {
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

	project, member, err := h.projectService.GetProjectDetails(c.Request.Context(), projectID, userID)
	if err != nil {
		if errors.Is(err, service.ErrProjectAccessDenied) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Project access denied")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectAccessDenied)))
			return
		}
		if errors.Is(err, service.ErrProjectNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectNotFound)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to get project details")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToProjectDetailResponse(project, member)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.UpdateProjectRequest
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

	// Update project
	project, err := h.projectService.UpdateProject(c.Request.Context(), projectID, userID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to update project")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrProjectNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectNotFound)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to update project")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Project updated")

	response := dto.ToProjectResponse(project)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
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

	err = h.projectService.DeleteProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to delete project")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to delete project")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Project deleted")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Project deleted successfully",
	}, nil))
}

// AddMember adds a member to the project
func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.AddMemberRequest
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

	targetUserID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	err = h.projectService.AddMember(c.Request.Context(), projectID, userID, targetUserID, req.Role, req.Permissions)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to add member")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrMemberAlreadyExists) {
			c.JSON(http.StatusConflict, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeMemberAlreadyExists)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
			Msg("Failed to add member")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
		Strs("permissions", req.Permissions).
		Msg("Member added to project")

	c.JSON(http.StatusCreated, dto.NewAPIResponse(map[string]string{
		"message": "Member added successfully",
	}, nil))
}

// GetMembers gets all members of a project with pagination
func (h *ProjectHandler) GetMembers(c *gin.Context) {
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

	members, totalCount, err := h.projectService.GetMembers(
		c.Request.Context(),
		projectID,
		userID,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		if errors.Is(err, service.ErrProjectAccessDenied) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Access denied to view members")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to get members")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Convert to responses with user details
	responses := make([]dto.ProjectMemberResponse, 0, len(members))
	for _, member := range members {
		user, err := h.userRepo.FindByID(c.Request.Context(), member.UserID)
		if err != nil {
			continue
		}
		responses = append(responses, dto.ToProjectMemberResponse(member, user))
	}

	paginationMeta := dto.NewPaginationMeta(params, totalCount)
	c.JSON(http.StatusOK, dto.NewAPIResponseWithPagination(responses, &paginationMeta))
}

// UpdateMember updates member permissions
func (h *ProjectHandler) UpdateMember(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	targetUserIDStr := c.Param("user_id")
	targetUserID, err := primitive.ObjectIDFromHex(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.UpdateMemberRequest
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

	err = h.projectService.UpdateMember(c.Request.Context(), projectID, userID, targetUserID, req.Role, req.Permissions)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to update member")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrMemberNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeMemberNotFound)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
			Msg("Failed to update member")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
		Strs("permissions", req.Permissions).
		Msg("Member permissions updated")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Member updated successfully",
	}, nil))
}

// RemoveMember removes a member from the project
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	targetUserIDStr := c.Param("user_id")
	targetUserID, err := primitive.ObjectIDFromHex(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	err = h.projectService.RemoveMember(c.Request.Context(), projectID, userID, targetUserID)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to remove member")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrCannotRemoveOwner) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
				Msg("Cannot remove last owner")
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeCannotRemoveOwner)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
			Msg("Failed to remove member")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Str("target_user_id", logger.SanitizeUserID(targetUserID.Hex())).
		Msg("Member removed from project")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Member removed successfully",
	}, nil))
}
