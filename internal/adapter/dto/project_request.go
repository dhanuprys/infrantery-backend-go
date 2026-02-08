package dto

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name                string `json:"name" validate:"required,min=1,max=100"`
	Description         string `json:"description" validate:"max=500"`
	EncryptionSalt      string `json:"encryption_salt" validate:"required"`
	EncryptedPrivateKey string `json:"encrypted_private_key" validate:"required"`
	EncryptionPublicKey string `json:"encryption_public_key" validate:"required"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// AddMemberRequest represents the request to add a member to a project
type AddMemberRequest struct {
	UserID      string   `json:"user_id" validate:"required"`
	Role        string   `json:"role" validate:"required,oneof=owner editor viewer custom"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,oneof=view_diagram edit_diagram view_note edit_note view_vault edit_vault manage_project"`
}

// UpdateMemberRequest represents the request to update member permissions
type UpdateMemberRequest struct {
	Role        string   `json:"role" validate:"required,oneof=owner editor viewer custom"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,oneof=view_diagram edit_diagram view_note edit_note view_vault edit_vault manage_project"`
}
