package dto

// CreateDiagramRequest represents a request to create a new diagram
type CreateDiagramRequest struct {
	DiagramName            string  `json:"diagram_name" validate:"required,min=1,max=255"`
	Description            string  `json:"description" validate:"omitempty,max=1000"`
	ParentDiagramID        *string `json:"parent_diagram_id,omitempty"`
	EncryptedData          *string `json:"encrypted_data,omitempty"`
	EncryptedDataSignature string  `json:"encrypted_data_signature" validate:"required"`
}

// UpdateDiagramRequest represents a request to update an existing diagram
type UpdateDiagramRequest struct {
	DiagramName            *string `json:"diagram_name,omitempty" validate:"omitempty,min=1,max=255"`
	Description            *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	EncryptedData          *string `json:"encrypted_data,omitempty"`
	EncryptedDataSignature *string `json:"encrypted_data_signature,omitempty"`
}
