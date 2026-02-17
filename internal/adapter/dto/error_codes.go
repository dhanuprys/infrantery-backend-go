package dto

// Error codes for the application
const (
	// Page Not Found errors
	ErrCodePageNotFound = "PAGE_NOT_FOUND"

	// Authentication errors
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	// Profile errors
	ErrCodeEmailAlreadyExists    = "EMAIL_ALREADY_EXISTS"
	ErrCodeUsernameAlreadyExists = "USERNAME_ALREADY_EXISTS"
	ErrCodeCurrentPasswordWrong  = "CURRENT_PASSWORD_WRONG"
	ErrCodeSamePassword          = "SAME_PASSWORD"

	// Project errors
	ErrCodeProjectNotFound        = "PROJECT_NOT_FOUND"
	ErrCodeProjectAccessDenied    = "PROJECT_ACCESS_DENIED"
	ErrCodeInsufficientPermission = "INSUFFICIENT_PERMISSION"
	ErrCodeMemberNotFound         = "MEMBER_NOT_FOUND"
	ErrCodeMemberAlreadyExists    = "MEMBER_ALREADY_EXISTS"
	ErrCodeCannotRemoveOwner      = "CANNOT_REMOVE_OWNER"

	// Invitation errors
	ErrCodeInvitationNotFound        = "INVITATION_NOT_FOUND"
	ErrCodeInvitationAlreadyAccepted = "INVITATION_ALREADY_ACCEPTED"
	ErrCodeInvitationExpired         = "INVITATION_EXPIRED"
	ErrCodeInvitationInvalidPassword = "INVITATION_INVALID_PASSWORD"

	// Note errors
	ErrCodeNoteNotFound     = "NOTE_NOT_FOUND"
	ErrCodeNoteAccessDenied = "NOTE_ACCESS_DENIED"
	ErrCodeInvalidNoteData  = "INVALID_NOTE_DATA"

	// Diagram errors
	ErrCodeDiagramNotFound     = "DIAGRAM_NOT_FOUND"
	ErrCodeDiagramAccessDenied = "DIAGRAM_ACCESS_DENIED"
	ErrCodeInvalidDiagramData  = "INVALID_DIAGRAM_DATA"

	// Node errors
	ErrCodeNodeNotFound     = "NODE_NOT_FOUND"
	ErrCodeNodeAccessDenied = "NODE_ACCESS_DENIED"
	ErrCodeInvalidNodeData  = "INVALID_NODE_DATA"
	ErrCodeInvalidNodeID    = "INVALID_NODE_ID"

	// Vault errors
	ErrCodeVaultItemNotFound    = "VAULT_ITEM_NOT_FOUND"
	ErrCodeVaultAccessDenied    = "VAULT_ACCESS_DENIED"
	ErrCodeInvalidVaultItemData = "INVALID_VAULT_ITEM_DATA"

	// Backup errors
	ErrCodeBackupTooLarge         = "BACKUP_TOO_LARGE"
	ErrCodeBackupInvalidFormat    = "BACKUP_INVALID_FORMAT"
	ErrCodeBackupVersionMismatch  = "BACKUP_VERSION_MISMATCH"
	ErrCodeBackupDecryptionFailed = "BACKUP_DECRYPTION_FAILED"

	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidRequest   = "INVALID_REQUEST"

	// Resource errors
	ErrCodeNotFound      = "RESOURCE_NOT_FOUND"
	ErrCodeAlreadyExists = "RESOURCE_ALREADY_EXISTS"
	ErrCodeForbidden     = "FORBIDDEN"

	// Server errors
	ErrCodeInternalError = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError = "DATABASE_ERROR"
)

// Error messages corresponding to error codes
var ErrorMessages = map[string]string{
	ErrCodePageNotFound: "Page not found",

	ErrCodeInvalidCredentials:     "Invalid email/username or password",
	ErrCodeUserAlreadyExists:      "User with this email or username already exists",
	ErrCodeInvalidToken:           "Invalid or expired token",
	ErrCodeExpiredToken:           "Token has expired",
	ErrCodeUnauthorized:           "Authorization required",
	ErrCodeEmailAlreadyExists:     "Email address is already in use",
	ErrCodeUsernameAlreadyExists:  "Username is already taken",
	ErrCodeCurrentPasswordWrong:   "Current password is incorrect",
	ErrCodeSamePassword:           "New password must be different from current password",
	ErrCodeProjectNotFound:        "Project not found",
	ErrCodeProjectAccessDenied:    "Access denied to this project",
	ErrCodeInsufficientPermission: "Insufficient permission to perform this action",
	ErrCodeMemberNotFound:         "Member not found",
	ErrCodeMemberAlreadyExists:    "Member already exists in this project",
	ErrCodeCannotRemoveOwner:      "Cannot remove the last owner from project",

	ErrCodeInvitationNotFound:        "Invitation not found",
	ErrCodeInvitationAlreadyAccepted: "Invitation has already been accepted",
	ErrCodeInvitationExpired:         "Invitation has expired",
	ErrCodeInvitationInvalidPassword: "Invalid invitation password",

	ErrCodeNoteNotFound:     "Note not found",
	ErrCodeNoteAccessDenied: "Access denied to this note",
	ErrCodeInvalidNoteData:  "Invalid note data provided",

	ErrCodeDiagramNotFound:     "Diagram not found",
	ErrCodeDiagramAccessDenied: "Access denied to this diagram",
	ErrCodeInvalidDiagramData:  "Invalid diagram data provided",

	ErrCodeNodeNotFound:     "Node not found",
	ErrCodeNodeAccessDenied: "Access denied to this node",
	ErrCodeInvalidNodeData:  "Invalid node data provided",
	ErrCodeInvalidNodeID:    "Invalid node ID format",

	ErrCodeVaultItemNotFound:    "Vault item not found",
	ErrCodeVaultAccessDenied:    "Access denied to this vault",
	ErrCodeInvalidVaultItemData: "Invalid vault item data provided",

	ErrCodeBackupTooLarge:         "Backup file exceeds maximum allowed size",
	ErrCodeBackupInvalidFormat:    "Invalid backup file format",
	ErrCodeBackupVersionMismatch:  "Unsupported backup version",
	ErrCodeBackupDecryptionFailed: "Decryption failed: wrong password or corrupted file",

	ErrCodeValidationFailed: "Validation failed",
	ErrCodeInvalidRequest:   "Invalid request body",
	ErrCodeNotFound:         "Resource not found",
	ErrCodeAlreadyExists:    "Resource already exists",
	ErrCodeForbidden:        "Access forbidden",
	ErrCodeInternalError:    "Internal server error",
	ErrCodeDatabaseError:    "Database operation failed",
}

// NewErrorResponse creates a new error response with code and message from dictionary
func NewErrorResponse(code string, customMessage ...string) *ErrorResponse {
	message := ErrorMessages[code]
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	}
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewValidationErrorResponse creates an error response for validation errors
func NewValidationErrorResponse(fields *[]map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Code:    ErrCodeValidationFailed,
		Message: ErrorMessages[ErrCodeValidationFailed],
		Fields:  fields,
	}
}
