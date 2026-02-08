package dto

type CreateNodeVaultRequest struct {
	Type                    string `json:"type" validate:"required"`
	EncryptedValue          string `json:"encrypted_value" validate:"required"`
	EncryptedValueSignature string `json:"encrypted_value_signature" validate:"required"`
}

type UpdateNodeVaultRequest struct {
	EncryptedValue          *string `json:"encrypted_value"`
	EncryptedValueSignature *string `json:"encrypted_value_signature"`
}

type NodeVaultResponse struct {
	ID                      string `json:"id"`
	NodeID                  string `json:"node_id"`
	ProjectID               string `json:"project_id"`
	Type                    string `json:"type"`
	EncryptedValue          string `json:"encrypted_value,omitempty"`
	EncryptedValueSignature string `json:"encrypted_value_signature,omitempty"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
}
