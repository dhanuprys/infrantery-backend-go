package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NodeVault struct {
	NodeId                  primitive.ObjectID `bson:"node_id" json:"node_id"`
	Type                    string             `bson:"type" json:"type"`
	EncryptedValue          *string            `bson:"encrypted_value,omitempty" json:"encrypted_value,omitempty"`
	EncryptedValueSignature *string            `bson:"encrypted_value_signature,omitempty" json:"encrypted_value_signature,omitempty"`
}
