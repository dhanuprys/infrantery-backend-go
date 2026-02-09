package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProjectID                 primitive.ObjectID `bson:"project_id" json:"project_id"`
	FileName                  string             `bson:"file_name" json:"file_name"`
	FileType                  string             `bson:"file_type" json:"file_type"`
	EncryptedContent          *string            `bson:"encrypted_content,omitempty" json:"encrypted_content,omitempty"`
	EncryptedContentSignature string             `bson:"encrypted_content_signature" json:"encrypted_content_signature"`

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}
