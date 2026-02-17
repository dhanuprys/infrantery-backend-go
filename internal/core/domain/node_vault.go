package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NodeVault struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NodeId primitive.ObjectID `bson:"node_id" json:"node_id"`
	// denormalized for performance on permission checking
	ProjectId               primitive.ObjectID `bson:"project_id" json:"project_id"`
	Label                   string             `bson:"label" json:"label"`
	Type                    string             `bson:"type" json:"type"`
	EncryptedValue          *string            `bson:"encrypted_value,omitempty" json:"encrypted_value,omitempty"`
	EncryptedValueSignature *string            `bson:"encrypted_value_signature,omitempty" json:"encrypted_value_signature,omitempty"`

	CreatedAt time.Time `bson:"createdAt,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updated_at"`
}
