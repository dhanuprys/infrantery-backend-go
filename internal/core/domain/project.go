package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`

	// Encryption keys for all project data (diagrams, notes, vaults)
	// encrypted + "<delimiter>" + salt + "<delimiter>" + iv
	SecretEncryptionPrivateKey string `bson:"secret_encrypted_private_key" json:"secret_encrypted_private_key"`
	EncryptionPublicKey        string `bson:"encryption_public_key" json:"encryption_public_key"`

	// Signing keys for all project data (diagrams, notes, vaults)
	// encrypted + "<delimiter>" + salt + "<delimiter" + iv
	SecretSigningPrivateKey string `bson:"secret_signing_private_key" json:"secret_signing_private_key"`
	SigningPublicKey        string `bson:"signing_public_key" json:"signing_public_key"`
}
