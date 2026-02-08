package handler

import (
	"net/http"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	validator   *validation.ValidationEngine
}

func NewAuthHandler(authService *service.AuthService, validator *validation.ValidationEngine) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} dto.APIResponse[dto.AuthResponse]
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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

	// Register user
	authResp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == service.ErrUserExists {
			logger.Warn().
				Str("email", logger.MaskEmail(req.Email)).
				Str("username", req.Username).
				Msg("Registration failed - user already exists")
			c.JSON(http.StatusConflict, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeUserAlreadyExists)))
			return
		}
		logger.Error().Err(err).Msg("Failed to register user")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("email", logger.MaskEmail(req.Email)).
		Msg("User registered successfully")

	c.JSON(http.StatusCreated, dto.NewAPIResponse(authResp, nil))
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.APIResponse[dto.AuthResponse]
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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

	// Login user
	authResp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			logger.Warn().
				Str("identifier", logger.MaskEmail(req.EmailOrUsername)).
				Msg("Login failed - invalid credentials")
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidCredentials)))
			return
		}
		logger.Error().Err(err).Msg("Login error")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("identifier", logger.MaskEmail(req.EmailOrUsername)).
		Msg("User logged in successfully")

	c.JSON(http.StatusOK, dto.NewAPIResponse(authResp, nil))
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} dto.APIResponse[dto.AuthResponse]
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
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

	// Refresh token
	authResp, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			logger.Warn().Msg("Token refresh failed - invalid or expired token")
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidToken)))
			return
		}
		logger.Error().Err(err).Msg("Failed to refresh token")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().Msg("Token refreshed successfully")

	c.JSON(http.StatusOK, dto.NewAPIResponse(authResp, nil))
}
