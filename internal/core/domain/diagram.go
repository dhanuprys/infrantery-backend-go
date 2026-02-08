package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Diagram struct {
	ID                     primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	ProjectID              primitive.ObjectID  `bson:"project_id" json:"project_id"`
	ParentDiagramID        *primitive.ObjectID `bson:"parent_diagram_id,omitempty" json:"parent_diagram_id,omitempty"`
	DiagramName            string              `bson:"diagram_name" json:"diagram_name"`
	Description            string              `bson:"description" json:"description"`
	EncryptedData          *string             `bson:"encrypted_data,omitempty" json:"encrypted_data,omitempty"`
	EncryptedDataSignature string              `bson:"encrypted_data_signature" json:"encrypted_data_signature"`
}
