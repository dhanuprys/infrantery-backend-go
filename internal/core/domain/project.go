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

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

type MemberKeyringUpdate struct {
	UserID              string
	EncryptedPassphrase string
	EncryptedSigningKey string
	SigningPublicKey    string
}
