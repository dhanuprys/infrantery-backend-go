package dto

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name                    string `json:"name" validate:"required,min=1,max=100"`
	Description             string `json:"description" validate:"max=500"`
	SecretPassphrase        string `json:"secret_passphrase" validate:"required"`
	SecretSigningPrivateKey string `json:"secret_signing_private_key" validate:"required"`
	SigningPublicKey        string `json:"signing_public_key" validate:"required"`
	UserPublicKey           string `json:"user_public_key" validate:"required"`
	UserEncryptedPrivateKey string `json:"user_encrypted_private_key" validate:"required"`
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

// CreateInvitationRequest represents the request to create an invitation
type CreateInvitationRequest struct {
	Role              string   `json:"role" validate:"required,oneof=owner editor viewer custom"`
	Permissions       []string `json:"permissions" validate:"required,min=1,dive,oneof=view_diagram edit_diagram view_note edit_note view_vault edit_vault manage_project"`
	InviteeUserID     string   `json:"invitee_user_id,omitempty" validate:"omitempty"`
	EncryptedKeyrings string   `json:"encrypted_keyrings" validate:"required"`
}

// AcceptInvitationRequest represents the request to accept an invitation
type AcceptInvitationRequest struct {
	Keyrings            []AcceptInvitationKeyring `json:"keyrings" validate:"required,min=1"`
	PublicKey           string                    `json:"public_key" validate:"required"`
	EncryptedPrivateKey string                    `json:"encrypted_private_key" validate:"required"`
}

// AcceptInvitationKeyring represents a keyring in the accept invitation request
type AcceptInvitationKeyring struct {
	Epoch                   string `json:"epoch" validate:"required"`
	SecretPassphrase        string `json:"secret_passphrase" validate:"required"`
	SecretSigningPrivateKey string `json:"secret_signing_private_key" validate:"required"`
	SigningPublicKey        string `json:"signing_public_key" validate:"required"`
}

// RotateProjectKeyRequest represents the request to rotate project keys
type RotateProjectKeyRequest struct {
	NewKeyEpoch string                `json:"new_key_epoch" validate:"required"`
	Updates     []MemberKeyringUpdate `json:"updates" validate:"required,min=1"`
}

// MemberKeyringUpdate represents the new keyring for a member
type MemberKeyringUpdate struct {
	UserID              string `json:"user_id" validate:"required"`
	EncryptedPassphrase string `json:"encrypted_passphrase" validate:"required"`
	EncryptedSigningKey string `json:"encrypted_signing_key" validate:"required"`
	SigningPublicKey    string `json:"signing_public_key" validate:"required"`
}
