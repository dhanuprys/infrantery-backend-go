package dto

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

// UserProfileResponse represents user profile information
type UserProfileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ToUserProfileResponse converts domain.User to UserProfileResponse
func ToUserProfileResponse(user *domain.User) *UserProfileResponse {
	return &UserProfileResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
