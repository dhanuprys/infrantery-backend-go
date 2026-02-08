package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`

	// Encryption keys for all project data (diagrams, notes, vaults)
	EncryptionSalt      string `bson:"encryption_salt" json:"encryption_salt"`
	EncryptedPrivateKey string `bson:"encrypted_private_key" json:"encrypted_private_key"`
	EncryptionPublicKey string `bson:"encryption_public_key" json:"encryption_public_key"`
}
