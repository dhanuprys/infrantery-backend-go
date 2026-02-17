package dto

// CreateBackupRequest is the request body for creating a backup.
type CreateBackupRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

// RestoreBackupResponse is the response after a successful restore.
type RestoreBackupResponse struct {
	Project ProjectResponse `json:"project"`
}
