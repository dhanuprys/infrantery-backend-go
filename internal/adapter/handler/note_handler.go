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

type NoteHandler struct {
	noteService *service.NoteService
	validator   *validation.ValidationEngine
}

func NewNoteHandler(
	noteService *service.NoteService,
	validator *validation.ValidationEngine,
) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
		validator:   validator,
	}
}

// CreateNote creates a new note in a project
func (h *NoteHandler) CreateNote(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.CreateNoteRequest
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

	// Parse ParentID if present
	var parentID *primitive.ObjectID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := primitive.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
			return
		}
		parentID = &pid
	}

	// Create note
	note, err := h.noteService.CreateNote(
		c.Request.Context(),
		projectID,
		userID,
		parentID,
		req.Type,
		req.FileName,
		req.Icon,
		req.EncryptedContent,
		&req.EncryptedContentSignature,
	)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("project_id", projectID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to create note")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrNoteAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to create note")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("note_id", note.ID.Hex()).
		Str("project_id", projectID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Note created")

	// TODO: Get timestamps from mgod
	response := dto.ToNoteResponse(note)
	c.JSON(http.StatusCreated, dto.NewAPIResponse(response, nil))
}

// ListNotes gets all notes for a project with pagination
func (h *NoteHandler) ListNotes(c *gin.Context) {
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

	notes, err := h.noteService.ListNotes(
		c.Request.Context(),
		projectID,
		userID,
	)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientPermission) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrNoteAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("project_id", projectID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to list notes")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// Convert to responses
	responses := make([]dto.NoteResponse, 0, len(notes))
	for _, note := range notes {
		// TODO: Get actual timestamps from mgod
		response := dto.ToNoteResponse(note)
		response.EncryptedContent = nil // Don't send content in list view
		response.EncryptedContentSignature = nil
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, dto.NewAPIResponse(responses, nil))
}

// GetNote gets a specific note
func (h *NoteHandler) GetNote(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	noteIDStr := c.Param("note_id")
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	note, err := h.noteService.GetNote(c.Request.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoteNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrNoteAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("note_id", noteID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to get note")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	// TODO: Get actual timestamps from mgod
	response := dto.ToNoteResponse(note)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// UpdateNote updates an existing note
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	noteIDStr := c.Param("note_id")
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	var req dto.UpdateNoteRequest
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

	// Update note
	note, err := h.noteService.UpdateNote(
		c.Request.Context(),
		noteID,
		userID,
		req.FileName,
		req.ParentID,
		req.Icon,
		req.EncryptedContent,
		req.EncryptedContentSignature,
	)
	if err != nil {
		if errors.Is(err, service.ErrNoteNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("note_id", noteID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to update note")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrNoteAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("note_id", noteID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to update note")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("note_id", noteID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Note updated")

	// TODO: Get actual timestamps from mgod
	response := dto.ToNoteResponse(note)
	c.JSON(http.StatusOK, dto.NewAPIResponse(response, nil))
}

// DeleteNote deletes a note
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	noteIDStr := c.Param("note_id")
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInvalidRequest)))
		return
	}

	// Get user ID from context
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	err = h.noteService.DeleteNote(c.Request.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoteNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteNotFound)))
			return
		}
		if errors.Is(err, service.ErrInsufficientPermission) {
			logger.Warn().
				Str("note_id", noteID.Hex()).
				Str("user_id", logger.SanitizeUserID(userID.Hex())).
				Msg("Insufficient permission to delete note")
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeInsufficientPermission)))
			return
		}
		if errors.Is(err, service.ErrNoteAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeNoteAccessDenied)))
			return
		}
		logger.Error().
			Err(err).
			Str("note_id", noteID.Hex()).
			Str("user_id", logger.SanitizeUserID(userID.Hex())).
			Msg("Failed to delete note")
		c.JSON(http.StatusInternalServerError, dto.NewAPIResponse[any](nil,
			dto.NewErrorResponse(dto.ErrCodeInternalError)))
		return
	}

	logger.Info().
		Str("note_id", noteID.Hex()).
		Str("user_id", logger.SanitizeUserID(userID.Hex())).
		Msg("Note deleted")

	c.JSON(http.StatusOK, dto.NewAPIResponse(map[string]string{
		"message": "Note deleted successfully",
	}, nil))
}
