package dto

// CreateNoteRequest represents a request to create a new note
type CreateNoteRequest struct {
	FileName                  string  `json:"file_name" validate:"required,min=1,max=255"`
	FileType                  string  `json:"file_type" validate:"required,oneof=markdown text"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature string  `json:"encrypted_content_signature" validate:"required"`
}

// UpdateNoteRequest represents a request to update an existing note
type UpdateNoteRequest struct {
	FileName                  *string `json:"file_name,omitempty" validate:"omitempty,min=1,max=255"`
	FileType                  *string `json:"file_type,omitempty" validate:"omitempty,oneof=markdown text"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature *string `json:"encrypted_content_signature,omitempty"`
}
