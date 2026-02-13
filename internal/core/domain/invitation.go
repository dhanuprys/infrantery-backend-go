package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusExpired  = "expired"
)

type Invitation struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProjectID         primitive.ObjectID `json:"project_id" bson:"project_id"`
	InviterUserID     primitive.ObjectID `json:"inviter_user_id" bson:"inviter_user_id"`
	InviteeUserID     primitive.ObjectID `json:"invitee_user_id,omitempty" bson:"invitee_user_id,omitempty"`
	Role              string             `json:"role" bson:"role"`
	Permissions       []string           `json:"permissions" bson:"permissions"`
	EncryptedKeyrings string             `json:"encrypted_keyrings" bson:"encrypted_keyrings"`
	KeyEpoch          string             `json:"key_epoch" bson:"key_epoch"`
	Status            string             `json:"status" bson:"status"`

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}
