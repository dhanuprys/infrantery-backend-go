package dto

// CreateNoteRequest represents a request to create a new note
type CreateNoteRequest struct {
	ParentID                  *string `json:"parent_id,omitempty" validate:"omitempty,len=24"`
	Type                      string  `json:"type" validate:"required,oneof=note folder"`
	FileName                  string  `json:"file_name" validate:"required,min=1,max=255"`
	Icon                      string  `json:"icon" validate:"omitempty,max=50"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature string  `json:"encrypted_content_signature" validate:"required_if=Type note"`
}

// UpdateNoteRequest represents a request to update an existing note
type UpdateNoteRequest struct {
	FileName                  *string `json:"file_name,omitempty" validate:"omitempty,min=1,max=255"`
	ParentID                  *string `json:"parent_id,omitempty" validate:"omitempty,len=24"`
	Icon                      *string `json:"icon,omitempty" validate:"omitempty,max=50"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature *string `json:"encrypted_content_signature,omitempty"`
}
