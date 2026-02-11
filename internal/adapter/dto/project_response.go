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
		Keyrings:    member.Keyrings,
		CreatedAt:   member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   member.UpdatedAt.Format(time.RFC3339),
	}
}
