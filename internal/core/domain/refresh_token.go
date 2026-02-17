package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	IsRevoked bool               `bson:"is_revoked" json:"is_revoked"`

	CreatedAt time.Time `bson:"createdAt,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updated_at"`
}
