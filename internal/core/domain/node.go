package domain

// Node is handled on frontend mostly, but if we need a schema:
type Node struct {
	ID                       string `bson:"id" json:"id"` // Frontend generate ID usually
	EncryptedReadme          string `bson:"encrypted_readme" json:"encrypted_readme"`
	EncryptedReadmeSignature string `bson:"encrypted_readme_signature" json:"encrypted_readme_signature"`
	EncryptedDict            string `bson:"encrypted_dict" json:"encrypted_dict"`
	EncryptedDictSignature   string `bson:"encrypted_dict_signature" json:"encrypted_dict_signature"`
}
