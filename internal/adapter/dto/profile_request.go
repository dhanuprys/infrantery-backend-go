package dto

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
}

// ChangePasswordRequest represents a request to change user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}
