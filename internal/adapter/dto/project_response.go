package dto

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

// ProjectResponse represents a basic project response
type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	KeyEpoch    string `json:"key_epoch"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProjectDetailResponse includes user's permissions
type ProjectDetailResponse struct {
	ID                      string                        `json:"id"`
	Name                    string                        `json:"name"`
	Description             string                        `json:"description"`
	KeyEpoch                string                        `json:"key_epoch"` // Changed from int64 to string
	Role                    string                        `json:"role"`
	Permissions             []string                      `json:"permissions"`
	UserEncryptedPrivateKey string                        `json:"user_encrypted_private_key"`
	Keyrings                []domain.ProjectMemberKeyring `json:"keyrings"`
	CreatedAt               string                        `json:"created_at"`
	UpdatedAt               string                        `json:"updated_at"`
}

// ProjectChunkResponse represents a project chunk
type ProjectChunkResponse struct {
	ID       string `json:"id"`
	KeyEpoch string `json:"key_epoch"`
}

// ProjectMemberResponse represents a project member
type ProjectMemberResponse struct {
	UserID      string                        `json:"user_id"`
	UserName    string                        `json:"user_name"`
	UserEmail   string                        `json:"user_email"`
	Role        string                        `json:"role"`
	Permissions []string                      `json:"permissions"`
	PublicKey   string                        `json:"public_key"`
	Keyrings    []domain.ProjectMemberKeyring `json:"keyrings"`
	CreatedAt   string                        `json:"created_at"`
	UpdatedAt   string                        `json:"updated_at"`
}

// ToProjectResponse converts a project to basic response
func ToProjectResponse(project *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID.Hex(),
		Name:        project.Name,
		Description: project.Description,
		KeyEpoch:    project.KeyEpoch,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
	}
}

// ToProjectDetailResponse converts a project and member to detailed response
func ToProjectDetailResponse(project *domain.Project, member *domain.ProjectMember) ProjectDetailResponse {
	return ProjectDetailResponse{
		ID:          project.ID.Hex(),
		Name:        project.Name,
		Description: project.Description,
		KeyEpoch:    project.KeyEpoch,
		Role:        member.Role,
		Permissions: member.Permissions,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
	}
}

func ToProjectChunkResponse(project *domain.Project) ProjectChunkResponse {
	return ProjectChunkResponse{
		ID:       project.ID.Hex(),
		KeyEpoch: project.KeyEpoch,
	}
}

// ToProjectMemberResponse converts member and user to response
func ToProjectMemberResponse(member *domain.ProjectMember, user *domain.User) ProjectMemberResponse {
	return ProjectMemberResponse{
		UserID:      member.UserID.Hex(),
		UserName:    user.Name,
		UserEmail:   user.Email,
		Role:        member.Role,
		Permissions: member.Permissions,
		PublicKey:   member.PublicKey,
		Keyrings:    member.Keyrings,
		CreatedAt:   member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   member.UpdatedAt.Format(time.RFC3339),
	}
}

// InvitationResponse represents an invitation
type InvitationResponse struct {
	ID                string   `json:"id"`
	ProjectID         string   `json:"project_id"`
	ProjectName       string   `json:"project_name"`
	InviterName       string   `json:"inviter_name"`
	InviteeName       string   `json:"invitee_name,omitempty"`
	Role              string   `json:"role"`
	Permissions       []string `json:"permissions"`
	EncryptedKeyrings string   `json:"encrypted_keyrings"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"created_at"`
}

// ToInvitationResponse converts an invitation to response
func ToInvitationResponse(invitation *domain.Invitation, projectName, inviterName, inviteeName string) InvitationResponse {
	return InvitationResponse{
		ID:                invitation.ID.Hex(),
		ProjectID:         invitation.ProjectID.Hex(),
		ProjectName:       projectName,
		InviterName:       inviterName,
		InviteeName:       inviteeName,
		Role:              invitation.Role,
		Permissions:       invitation.Permissions,
		EncryptedKeyrings: invitation.EncryptedKeyrings,
		Status:            invitation.Status,
		CreatedAt:         invitation.CreatedAt.Format(time.RFC3339),
	}
}

// UserSearchResponse represents a user search result
type UserSearchResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// ToUserSearchResponse converts a user to search response
func ToUserSearchResponse(user *domain.User) UserSearchResponse {
	return UserSearchResponse{
		ID:       user.ID.Hex(),
		Name:     user.Name,
		Email:    user.Email,
		Username: user.Username,
	}
}
