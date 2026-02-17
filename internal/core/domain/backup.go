package domain

import "time"

// BackupVersion is the current backup format version.
const BackupVersion = 1

// BackupMagic is the magic header bytes for backup files.
var BackupMagic = []byte("INFBK")

// BackupPepper is a hardcoded application secret mixed into the key
// derivation via HMAC. This ensures backup files can only be decrypted
// by this application — even if a third party knows the user's password,
// they cannot derive the correct encryption key without this pepper.
var BackupPepper = []byte("infrantery:backup:v1:a9f2c8e1-4d7b-4f3a-b5e6-8c1d9e0f7a2b")

// BackupPayload is the top-level structure serialized to JSON
// before compression and encryption.
type BackupPayload struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	Project   ProjectBackup   `json:"project"`
	Member    MemberBackup    `json:"member"`
	Diagrams  []DiagramBackup `json:"diagrams"`
	Nodes     []NodeBackup    `json:"nodes"`
	Vaults    []VaultBackup   `json:"vaults"`
	Notes     []NoteBackup    `json:"notes"`
}

// ProjectBackup is the portable representation of a Project.
type ProjectBackup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	KeyEpoch    string `json:"key_epoch"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// MemberBackup stores the backup creator's member record so the
// restoring user can immediately access encrypted content.
type MemberBackup struct {
	PublicKey           string                `json:"public_key"`
	EncryptedPrivateKey string                `json:"encrypted_private_key"`
	Keyrings            []MemberKeyringBackup `json:"keyrings"`
}

// MemberKeyringBackup is a portable copy of ProjectMemberKeyring.
type MemberKeyringBackup struct {
	Epoch                   string `json:"epoch"`
	SecretPassphrase        string `json:"secret_passphrase"`
	SecretSigningPrivateKey string `json:"secret_signing_private_key"`
	SigningPublicKey        string `json:"signing_public_key"`
}

// DiagramBackup is the portable representation of a Diagram.
type DiagramBackup struct {
	ID                     string  `json:"id"`
	ParentDiagramID        *string `json:"parent_diagram_id,omitempty"`
	DiagramName            string  `json:"diagram_name"`
	Description            string  `json:"description"`
	EncryptedData          *string `json:"encrypted_data,omitempty"`
	EncryptedDataSignature string  `json:"encrypted_data_signature"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

// NodeBackup is the portable representation of a Node.
type NodeBackup struct {
	ID                       string `json:"id"`
	DiagramID                string `json:"diagram_id"`
	EncryptedReadme          string `json:"encrypted_readme"`
	EncryptedReadmeSignature string `json:"encrypted_readme_signature"`
	EncryptedDict            string `json:"encrypted_dict"`
	EncryptedDictSignature   string `json:"encrypted_dict_signature"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

// VaultBackup is the portable representation of a NodeVault.
type VaultBackup struct {
	ID                      string  `json:"id"`
	NodeID                  string  `json:"node_id"`
	Label                   string  `json:"label"`
	Type                    string  `json:"type"`
	EncryptedValue          *string `json:"encrypted_value,omitempty"`
	EncryptedValueSignature *string `json:"encrypted_value_signature,omitempty"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}

// NoteBackup is the portable representation of a Note.
type NoteBackup struct {
	ID                        string  `json:"id"`
	ParentID                  *string `json:"parent_id,omitempty"`
	Type                      string  `json:"type"`
	FileName                  string  `json:"file_name"`
	Icon                      string  `json:"icon,omitempty"`
	EncryptedContent          *string `json:"encrypted_content,omitempty"`
	EncryptedContentSignature *string `json:"encrypted_content_signature,omitempty"`
	CreatedAt                 string  `json:"created_at"`
	UpdatedAt                 string  `json:"updated_at"`
}
