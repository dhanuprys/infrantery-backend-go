package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Node represents a node in a diagram with encrypted extended data
type Node struct {
	ID                       primitive.ObjectID `bson:"_id" json:"id"`
	DiagramID                primitive.ObjectID `bson:"diagram_id" json:"diagram_id"`
	EncryptedReadme          string             `bson:"encrypted_readme" json:"encrypted_readme"`
	EncryptedReadmeSignature string             `bson:"encrypted_readme_signature" json:"encrypted_readme_signature"`
	EncryptedDict            string             `bson:"encrypted_dict" json:"encrypted_dict"`
	EncryptedDictSignature   string             `bson:"encrypted_dict_signature" json:"encrypted_dict_signature"`

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}
