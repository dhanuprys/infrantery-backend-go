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

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}
