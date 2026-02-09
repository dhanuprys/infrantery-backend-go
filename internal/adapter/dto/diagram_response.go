package dto

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

// DiagramResponse represents a diagram in API responses
type DiagramResponse struct {
	ID                     string  `json:"id"`
	ProjectID              string  `json:"project_id"`
	ParentDiagramID        *string `json:"parent_diagram_id,omitempty"`
	DiagramName            string  `json:"diagram_name"`
	Description            string  `json:"description"`
	EncryptedData          *string `json:"encrypted_data,omitempty"`
	EncryptedDataSignature string  `json:"encrypted_data_signature"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

// ToDiagramResponse converts a domain Diagram to DiagramResponse
func ToDiagramResponse(diagram *domain.Diagram) DiagramResponse {
	var parentID *string
	if diagram.ParentDiagramID != nil {
		hexID := diagram.ParentDiagramID.Hex()
		parentID = &hexID
	}

	return DiagramResponse{
		ID:                     diagram.ID.Hex(),
		ProjectID:              diagram.ProjectID.Hex(),
		ParentDiagramID:        parentID,
		DiagramName:            diagram.DiagramName,
		Description:            diagram.Description,
		EncryptedData:          diagram.EncryptedData,
		EncryptedDataSignature: diagram.EncryptedDataSignature,
		CreatedAt:              diagram.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              diagram.UpdatedAt.Format(time.RFC3339),
	}
}
