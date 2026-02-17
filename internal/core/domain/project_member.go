package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionViewDiagram   = "view_diagram"
	PermissionEditDiagram   = "edit_diagram"
	PermissionViewNote      = "view_note"
	PermissionEditNote      = "edit_note"
	PermissionViewVault     = "view_vault"
	PermissionEditVault     = "edit_vault"
	PermissionManageProject = "manage_project"
)

type ProjectMember struct {
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Permissions []string           `bson:"permissions" json:"permissions"`
	Role        string             `bson:"role" json:"role"` // Optional preset name

	// User key pair
	PublicKey string `bson:"public_key" json:"public_key"`
	// encrypted + "<delimiter>" + salt + "<delimiter>" + iv
	EncryptedPrivateKey string `bson:"encrypted_private_key" json:"encrypted_private_key"`

	Keyrings []ProjectMemberKeyring `bson:"keyrings,omitempty" json:"keyrings"`

	CreatedAt time.Time `bson:"createdAt,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updated_at"`
}

type ProjectMemberKeyring struct {
	Epoch string `bson:"epoch" json:"epoch"`

	// Encryption keys for all project data (diagrams, notes, vaults)
	// encrypted + "<delimiter>" + salt + "<delimiter>" + iv
	SecretPassphrase string `bson:"secret_passphrase" json:"secret_passphrase"`

	// Signing keys for all project data (diagrams, notes, vaults)
	// encrypted + "<delimiter>" + salt + "<delimiter" + iv
	SecretSigningPrivateKey string `bson:"secret_signing_private_key" json:"secret_signing_private_key"`
	SigningPublicKey        string `bson:"signing_public_key" json:"signing_public_key"`
}
