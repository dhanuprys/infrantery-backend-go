package dto

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

// NoteResponse represents a note in API responses
type NoteResponse struct {
	ID                        string  `json:"id"`
	ProjectID                 string  `json:"project_id"`
	ParentID                  *string `json:"parent_id,omitempty"`
	Type                      string  `json:"type"`
	FileName                  string  `json:"file_name"`
	Icon                      string  `json:"icon"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature *string `json:"encrypted_content_signature,omitempty"`
	CreatedAt                 string  `json:"created_at"`
	UpdatedAt                 string  `json:"updated_at"`
}

// ToNoteResponse converts a domain Note to NoteResponse
func ToNoteResponse(note *domain.Note) NoteResponse {
	response := NoteResponse{
		ID:                        note.ID.Hex(),
		ProjectID:                 note.ProjectID.Hex(),
		FileName:                  note.FileName,
		Type:                      note.Type,
		Icon:                      note.Icon,
		EncryptedContent:          note.EncryptedContent,
		EncryptedContentSignature: note.EncryptedContentSignature,
		CreatedAt:                 note.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 note.UpdatedAt.Format(time.RFC3339),
	}

	if note.ParentID != nil {
		parentID := note.ParentID.Hex()
		response.ParentID = &parentID
	}

	return response
}
