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

// BackupHandler handles HTTP requests for backup and restore operations.
type BackupHandler struct {
	backupService *service.BackupService
	validator     *validation.ValidationEngine
}

// NewBackupHandler creates a new BackupHandler.
func NewBackupHandler(
	backupService *service.BackupService,
	validator *validation.ValidationEngine,
) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		validator:     validator,
	}
}

// CreateBackup handles POST /projects/:project_id/backup
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	var req dto.CreateBackupRequest
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

	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Invalid project ID")))
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	reader, filename, err := h.backupService.CreateBackup(c.Request.Context(), projectID, userID, req.Password)
	if err != nil {
		logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to create backup")

		if errors.Is(err, service.ErrBackupTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeBackupTooLarge)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrProjectAccessDenied) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeProjectAccessDenied)))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// RestoreBackup handles POST /projects/restore
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Backup file is required")))
		return
	}

	password := c.PostForm("password")
	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Password must be at least 8 characters")))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest, "Cannot read backup file")))
		return
	}
	defer file.Close()

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	project, err := h.backupService.RestoreBackup(c.Request.Context(), userID, password, file)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to restore backup")

		switch {
		case errors.Is(err, service.ErrBackupTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeBackupTooLarge)))
		case errors.Is(err, service.ErrBackupInvalidFormat):
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeBackupInvalidFormat)))
		case errors.Is(err, service.ErrBackupVersionMismatch):
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeBackupVersionMismatch)))
		case errors.Is(err, service.ErrBackupDecryptionFailed):
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeBackupDecryptionFailed)))
		default:
			c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInternalError)))
		}
		return
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(
		&dto.RestoreBackupResponse{
			Project: dto.ToProjectResponse(project),
		},
		nil,
	))
}
