package dto

type GetOrCreateNodeRequest struct {
	NodeID string `json:"node_id" validate:"required,mongo_id"` // Must be valid ObjectID hex
}

type UpdateNodeRequest struct {
	EncryptedReadme          *string `json:"encrypted_readme,omitempty"`
	EncryptedReadmeSignature *string `json:"encrypted_readme_signature,omitempty"`
	EncryptedDict            *string `json:"encrypted_dict,omitempty"`
	EncryptedDictSignature   *string `json:"encrypted_dict_signature,omitempty"`
}
