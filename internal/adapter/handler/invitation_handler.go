package handler

import (
	"errors"
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InvitationHandler struct {
	projectService *service.ProjectService
	userRepo       port.UserRepository
	projectRepo    port.ProjectRepository
	validator      *validation.ValidationEngine
}

func NewInvitationHandler(
	projectService *service.ProjectService,
	userRepo port.UserRepository,
	projectRepo port.ProjectRepository,
	validator *validation.ValidationEngine,
) *InvitationHandler {
	return &InvitationHandler{
		projectService: projectService,
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		validator:      validator,
	}
}

// GetInvitation fetches an invitation by ID (for invitee)
func (h *InvitationHandler) GetInvitation(c *gin.Context) {
	invitationIDStr := c.Param("invitation_id")
	invitationID, err := primitive.ObjectIDFromHex(invitationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	invitation, err := h.projectService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		if errors.Is(err, service.ErrInvitationNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvitationNotFound)))
			return
		}
		logger.Error().Err(err).Str("invitation_id", invitationIDStr).Msg("Failed to get invitation")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Get project and inviter names for response
	project, err := h.projectRepo.FindByID(c.Request.Context(), invitation.ProjectID)
	projectName := ""
	if err == nil && project != nil {
		projectName = project.Name
	}

	inviter, err := h.userRepo.FindByID(c.Request.Context(), invitation.InviterUserID)
	inviterName := ""
	if err == nil && inviter != nil {
		inviterName = inviter.Name
	}

	inviteeName := ""
	if !invitation.InviteeUserID.IsZero() {
		invitee, _ := h.userRepo.FindByID(c.Request.Context(), invitation.InviteeUserID)
		if invitee != nil {
			inviteeName = invitee.Name
		}
	}

	response := dto.ToInvitationResponse(invitation, projectName, inviterName, inviteeName)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// AcceptInvitation accepts an invitation (for invitee)
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	invitationIDStr := c.Param("invitation_id")
	invitationID, err := primitive.ObjectIDFromHex(invitationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	if validationErrors := h.validator.ValidateStruct(req); validationErrors != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewValidationErrorResponse(validationErrors)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Convert DTO keyrings to domain keyrings
	keyrings := make([]domain.ProjectMemberKeyring, len(req.Keyrings))
	for i, kr := range req.Keyrings {
		keyrings[i] = domain.ProjectMemberKeyring{
			Epoch:                   kr.Epoch,
			SecretPassphrase:        kr.SecretPassphrase,
			SecretSigningPrivateKey: kr.SecretSigningPrivateKey,
			SigningPublicKey:        kr.SigningPublicKey,
		}
	}

	projectID, err := h.projectService.AcceptInvitation(
		c.Request.Context(),
		invitationID,
		userID,
		keyrings,
		req.PublicKey,
		req.EncryptedPrivateKey,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvitationNotFound) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvitationNotFound)))
			return
		}
		if errors.Is(err, service.ErrInvitationAlreadyAccepted) {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvitationAlreadyAccepted)))
			return
		}
		if errors.Is(err, service.ErrInvitationExpired) {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvitationExpired)))
			return
		}
		if errors.Is(err, service.ErrInvitationInvalidPassword) {
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvitationInvalidPassword)))
			return
		}
		if errors.Is(err, service.ErrMemberAlreadyExists) {
			c.JSON(http.StatusConflict, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeMemberAlreadyExists)))
			return
		}
		logger.Error().Err(err).
			Str("invitation_id", invitationIDStr).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to accept invitation")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("invitation_id", invitationIDStr).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Invitation accepted")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message":    "Invitation accepted successfully",
		"project_id": projectID.Hex(),
	}, nil))
}

// SearchUsers searches for users by name, email, or username
func (h *InvitationHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, dto.NewAPIResponse([]dto.UserSearchResponse{}, nil))
		return
	}

	// Get current user ID to exclude from results
	userIDStr, _ := c.Get("user_id")
	currentUserID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	users, err := h.userRepo.SearchUsers(c.Request.Context(), query, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to search users")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Filter out current user
	responses := make([]dto.UserSearchResponse, 0, len(users))
	for _, user := range users {
		if user.ID != currentUserID {
			responses = append(responses, dto.ToUserSearchResponse(user))
		}
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(responses, nil))
}

// ListUserInvitations lists invitations for the current user
func (h *InvitationHandler) ListUserInvitations(c *gin.Context) {
	// Get current user ID
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Parse query params for pagination
	params := dto.DefaultPaginationParams()
	if err := c.ShouldBindQuery(&params); err != nil {
		// Ignore error, use defaults
	}
	params.Validate()

	invitations, total, err := h.projectService.GetUserInvitations(c.Request.Context(), userID, params.GetOffset(), params.GetLimit())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list user invitations")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	responses := make([]dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		// Fetch project name
		projectName := "Unknown Project"
		project, err := h.projectRepo.FindByID(c.Request.Context(), inv.ProjectID)
		if err == nil && project != nil {
			projectName = project.Name
		}

		// Fetch inviter name
		inviterName := "Unknown User"
		inviter, err := h.userRepo.FindByID(c.Request.Context(), inv.InviterUserID)
		if err == nil && inviter != nil {
			inviterName = inviter.Name
		}

		// Invitee name is current user
		inviteeName := "" // Optional to fill

		responses = append(responses, dto.ToInvitationResponse(inv, projectName, inviterName, inviteeName))
	}

	metadata := dto.NewPaginationMeta(params, total)
	c.JSON(http.StatusOK, dto.NewAPIResponseWithPagination(responses, &metadata))
}
