package dto

import "github.com/dhanuprys/infrantery-backend-go/internal/core/domain"

type NodeResponse struct {
	ID                       string `json:"id"`
	DiagramID                string `json:"diagram_id"`
	EncryptedReadme          string `json:"encrypted_readme,omitempty"`
	EncryptedReadmeSignature string `json:"encrypted_readme_signature,omitempty"`
	EncryptedDict            string `json:"encrypted_dict,omitempty"`
	EncryptedDictSignature   string `json:"encrypted_dict_signature,omitempty"`
}

func ToNodeResponse(node *domain.Node) NodeResponse {
	return NodeResponse{
		ID:                       node.ID.Hex(),
		DiagramID:                node.DiagramID.Hex(),
		EncryptedReadme:          node.EncryptedReadme,
		EncryptedReadmeSignature: node.EncryptedReadmeSignature,
		EncryptedDict:            node.EncryptedDict,
		EncryptedDictSignature:   node.EncryptedDictSignature,
	}
}
