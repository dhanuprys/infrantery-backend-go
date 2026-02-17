package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`

	KeyEpoch string `bson:"key_epoch" json:"key_epoch"`

	CreatedAt time.Time `bson:"createdAt,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updated_at"`
}

type MemberKeyringUpdate struct {
	UserID              string
	EncryptedPassphrase string
	EncryptedSigningKey string
	SigningPublicKey    string
}
