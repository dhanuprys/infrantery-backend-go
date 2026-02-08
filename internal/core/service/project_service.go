package service

import (
	"context"
	"errors"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProjectNotFound        = errors.New("project not found")
	ErrProjectAccessDenied    = errors.New("project access denied")
	ErrInsufficientPermission = errors.New("insufficient permission")
	ErrMemberNotFound         = errors.New("member not found")
	ErrMemberAlreadyExists    = errors.New("member already exists")
	ErrCannotRemoveOwner      = errors.New("cannot remove last owner")
)

// RolePresets defines default permissions for each role
var RolePresets = map[string][]string{
	"owner": {
		domain.PermissionViewDiagram, domain.PermissionEditDiagram,
		domain.PermissionViewNote, domain.PermissionEditNote,
		domain.PermissionViewVault, domain.PermissionEditVault,
		domain.PermissionManageProject,
	},
	"editor": {
		domain.PermissionViewDiagram, domain.PermissionEditDiagram,
		domain.PermissionViewNote, domain.PermissionEditNote,
		domain.PermissionViewVault,
	},
	"viewer": {
		domain.PermissionViewDiagram,
		domain.PermissionViewNote,
	},
}

type ProjectService struct {
	projectRepo port.ProjectRepository
	memberRepo  port.ProjectMemberRepository
	userRepo    port.UserRepository
	noteRepo    port.NoteRepository
	diagramRepo port.DiagramRepository
}

func NewProjectService(
	projectRepo port.ProjectRepository,
	memberRepo port.ProjectMemberRepository,
	userRepo port.UserRepository,
	noteRepo port.NoteRepository,
	diagramRepo port.DiagramRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		noteRepo:    noteRepo,
		diagramRepo: diagramRepo,
	}
}

// CreateProject creates a new project with the creator as owner
func (s *ProjectService) CreateProject(
	ctx context.Context,
	userID primitive.ObjectID,
	name, description string,
	encryptionSalt, encryptedPrivateKey, encryptionPublicKey string,
) (*domain.Project, error) {
	project := &domain.Project{
		ID:                  primitive.NewObjectID(),
		Name:                name,
		Description:         description,
		EncryptionSalt:      encryptionSalt,
		EncryptedPrivateKey: encryptedPrivateKey,
		EncryptionPublicKey: encryptionPublicKey,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	// Add creator as owner
	member := &domain.ProjectMember{
		ProjectID:   project.ID,
		UserID:      userID,
		Role:        "owner",
		Permissions: RolePresets["owner"],
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return project, nil
}

// GetUserProjects gets all projects the user has access to with pagination
func (s *ProjectService) GetUserProjects(ctx context.Context, userID primitive.ObjectID, offset, limit int) ([]*domain.Project, int64, error) {
	return s.projectRepo.FindByUserID(ctx, userID, offset, limit)
}

// GetProjectDetails gets project details with user permissions
func (s *ProjectService) GetProjectDetails(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
) (*domain.Project, *domain.ProjectMember, error) {
	// Check if user has access
	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil, ErrProjectAccessDenied
		}
		return nil, nil, err
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil, ErrProjectNotFound
		}
		return nil, nil, err
	}

	return project, member, nil
}

// UpdateProject updates project details
func (s *ProjectService) UpdateProject(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	name, description *string,
) (*domain.Project, error) {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return nil, err
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if name != nil {
		project.Name = *name
	}
	if description != nil {
		project.Description = *description
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject deletes a project (owner only)
func (s *ProjectService) DeleteProject(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
) error {
	// Check permission (only owners can delete)
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	// Cascade delete: Delete all members first
	if err := s.memberRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return err
	}

	// Cascade delete: Delete all notes
	if err := s.noteRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return err
	}

	// Cascade delete: Delete all diagrams
	if err := s.diagramRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return err
	}

	// Delete the project
	return s.projectRepo.Delete(ctx, projectID)
}

// AddMember adds a member to the project
func (s *ProjectService) AddMember(
	ctx context.Context,
	projectID, userID, targetUserID primitive.ObjectID,
	role string,
	permissions []string,
) error {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	// Check if target user exists
	_, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrMemberNotFound
		}
		return err
	}

	// Check if member already exists
	_, err = s.memberRepo.FindByProjectAndUser(ctx, projectID, targetUserID)
	if err == nil {
		return ErrMemberAlreadyExists
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	// Create member
	member := &domain.ProjectMember{
		ProjectID:   projectID,
		UserID:      targetUserID,
		Role:        role,
		Permissions: permissions,
	}

	return s.memberRepo.Create(ctx, member)
}

// GetMembers gets all members of a project with pagination
func (s *ProjectService) GetMembers(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	offset, limit int,
) ([]*domain.ProjectMember, int64, error) {
	// Check if user has access (any member can view members)
	_, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, 0, ErrProjectAccessDenied
		}
		return nil, 0, err
	}

	return s.memberRepo.FindByProjectID(ctx, projectID, offset, limit)
}

// UpdateMember updates member permissions
func (s *ProjectService) UpdateMember(
	ctx context.Context,
	projectID, userID, targetUserID primitive.ObjectID,
	role string,
	permissions []string,
) error {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, targetUserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrMemberNotFound
		}
		return err
	}

	member.Role = role
	member.Permissions = permissions

	return s.memberRepo.Update(ctx, member)
}

// RemoveMember removes a member from the project
func (s *ProjectService) RemoveMember(
	ctx context.Context,
	projectID, userID, targetUserID primitive.ObjectID,
) error {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	// Check if target is the last owner (fetch all members to count owners)
	members, _, err := s.memberRepo.FindByProjectID(ctx, projectID, 0, 10000) // Get all members
	if err != nil {
		return err
	}

	ownerCount := 0
	targetIsOwner := false
	for _, m := range members {
		if m.Role == "owner" {
			ownerCount++
			if m.UserID == targetUserID {
				targetIsOwner = true
			}
		}
	}

	if targetIsOwner && ownerCount == 1 {
		return ErrCannotRemoveOwner
	}

	return s.memberRepo.Delete(ctx, projectID, targetUserID)
}

// HasPermission checks if user has a specific permission
func (s *ProjectService) HasPermission(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	permission string,
) error {
	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrProjectAccessDenied
		}
		return err
	}

	for _, p := range member.Permissions {
		if p == permission {
			return nil
		}
	}

	return ErrInsufficientPermission
}

// GetUserPermissions gets user's permissions for a project
func (s *ProjectService) GetUserPermissions(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
) ([]string, error) {
	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProjectAccessDenied
		}
		return nil, err
	}

	return member.Permissions, nil
}
