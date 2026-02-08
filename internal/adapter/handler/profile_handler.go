package handler

import (
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileHandler struct {
	userService *service.UserService
	validator   *validation.ValidationEngine
}

func NewProfileHandler(userService *service.UserService, validator *validation.ValidationEngine) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
		validator:   validator,
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags profile
// @Produce json
// @Success 200 {object} dto.APIResponse[dto.UserProfileResponse]
// @Router /api/v1/profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Get user ID from auth middleware context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeUnauthorized)))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid user ID")))
		return
	}

	// Get user profile
	user, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNotFound, "User not found")))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	response := dto.ToUserProfileResponse(user.ID, user.Name, user.Username, user.Email)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// UpdateProfile godoc
// @Summary Update user profile
// @Tags profile
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} dto.APIResponse[dto.UserProfileResponse]
// @Router /api/v1/profile [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from auth middleware context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeUnauthorized)))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid user ID")))
		return
	}

	var req dto.UpdateProfileRequest
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

	// Update profile
	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		if err == service.ErrEmailAlreadyExists {
			logger.Warn().
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Profile update failed - email already exists")
			c.JSON(http.StatusConflict, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeEmailAlreadyExists)))
			return
		}
		if err == service.ErrUsernameAlreadyExists {
			logger.Warn().
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Profile update failed - username already exists")
			c.JSON(http.StatusConflict, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeUsernameAlreadyExists)))
			return
		}
		logger.Error().
			Err(err).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to update profile")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Profile updated successfully")

	response := dto.ToUserProfileResponse(user.ID, user.Name, user.Username, user.Email)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// ChangePassword godoc
// @Summary Change user password
// @Tags profile
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Change Password Request"
// @Success 200 {object} dto.APIResponse[any]
// @Router /api/v1/profile/password [put]
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	// Get user ID from auth middleware context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeUnauthorized)))
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid user ID")))
		return
	}

	var req dto.ChangePasswordRequest
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

	// Change password
	err = h.userService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err == service.ErrCurrentPasswordWrong {
			logger.Warn().
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Password change failed - incorrect current password")
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeCurrentPasswordWrong)))
			return
		}
		if err == service.ErrSamePassword {
			logger.Warn().
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Password change failed - same password")
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeSamePassword)))
			return
		}
		logger.Error().
			Err(err).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to change password")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Password changed successfully")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Password changed successfully",
	}, nil))
}
