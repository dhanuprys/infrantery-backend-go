package dto

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

type CreateNodeVaultRequest struct {
	Label                   string `json:"label" validate:"required"`
	Type                    string `json:"type" validate:"required"`
	EncryptedValue          string `json:"encrypted_value" validate:"required"`
	EncryptedValueSignature string `json:"encrypted_value_signature" validate:"required"`
}

type UpdateNodeVaultRequest struct {
	Label                   *string `json:"label"`
	EncryptedValue          *string `json:"encrypted_value"`
	EncryptedValueSignature *string `json:"encrypted_value_signature"`
}

type NodeVaultResponse struct {
	ID                      string `json:"id"`
	NodeID                  string `json:"node_id"`
	ProjectID               string `json:"project_id"`
	Label                   string `json:"label"`
	Type                    string `json:"type"`
	EncryptedValue          string `json:"encrypted_value,omitempty"`
	EncryptedValueSignature string `json:"encrypted_value_signature,omitempty"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
}

func ToNodeVaultResponse(vault *domain.NodeVault) NodeVaultResponse {
	return NodeVaultResponse{
		ID:        vault.ID.Hex(),
		NodeID:    vault.NodeId.Hex(),
		ProjectID: vault.ProjectId.Hex(),
		Label:     vault.Label,
		Type:      vault.Type,
		EncryptedValue: func() string {
			if vault.EncryptedValue != nil {
				return *vault.EncryptedValue
			}
			return ""
		}(),
		EncryptedValueSignature: func() string {
			if vault.EncryptedValueSignature != nil {
				return *vault.EncryptedValueSignature
			}
			return ""
		}(),
		CreatedAt: vault.CreatedAt.Format(time.RFC3339),
		UpdatedAt: vault.UpdatedAt.Format(time.RFC3339),
	}
}
