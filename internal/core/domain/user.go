package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"-"` // Never return password in JSON
	Email    string             `bson:"email" json:"email"`

	CreatedAt time.Time `bson:"createdAt,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updated_at"`
}
