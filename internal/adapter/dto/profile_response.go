package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

// UserProfileResponse represents user profile information
type UserProfileResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ToUserProfileResponse converts domain.User to UserProfileResponse
func ToUserProfileResponse(id primitive.ObjectID, name, username, email string) *UserProfileResponse {
	return &UserProfileResponse{
		ID:       id.Hex(),
		Name:     name,
		Username: username,
		Email:    email,
	}
}
