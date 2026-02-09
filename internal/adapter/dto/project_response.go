package dto

import (
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
)

// ProjectResponse represents a basic project response
type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProjectDetailResponse includes user's permissions
type ProjectDetailResponse struct {
	ID                         string   `json:"id"`
	Name                       string   `json:"name"`
	Description                string   `json:"description"`
	SecretEncryptionPrivateKey string   `json:"secret_encrypted_private_key"`
	EncryptionPublicKey        string   `json:"encryption_public_key"`
	SecretSigningPrivateKey    string   `json:"secret_signing_private_key"`
	SigningPublicKey           string   `json:"signing_public_key"`
	Role                       string   `json:"role"`
	Permissions                []string `json:"permissions"`
}

// ProjectMemberResponse represents a project member
type ProjectMemberResponse struct {
	UserID      string   `json:"user_id"`
	UserName    string   `json:"user_name"`
	UserEmail   string   `json:"user_email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// ToProjectResponse converts a project to basic response
func ToProjectResponse(project *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID.Hex(),
		Name:        project.Name,
		Description: project.Description,
	}
}

// ToProjectDetailResponse converts a project and member to detailed response
func ToProjectDetailResponse(project *domain.Project, member *domain.ProjectMember) ProjectDetailResponse {
	return ProjectDetailResponse{
		ID:                         project.ID.Hex(),
		Name:                       project.Name,
		Description:                project.Description,
		SecretEncryptionPrivateKey: project.SecretEncryptionPrivateKey,
		EncryptionPublicKey:        project.EncryptionPublicKey,
		SecretSigningPrivateKey:    project.SecretSigningPrivateKey,
		SigningPublicKey:           project.SigningPublicKey,
		Role:                       member.Role,
		Permissions:                member.Permissions,
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
	}
}
