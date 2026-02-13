package service

import (
	"context"
	"errors"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProjectNotFound           = errors.New("project not found")
	ErrProjectAccessDenied       = errors.New("project access denied")
	ErrInsufficientPermission    = errors.New("insufficient permission")
	ErrMemberNotFound            = errors.New("member not found")
	ErrMemberAlreadyExists       = errors.New("member already exists")
	ErrCannotRemoveOwner         = errors.New("cannot remove last owner")
	ErrInvitationNotFound        = errors.New("invitation not found")
	ErrInvitationAlreadyAccepted = errors.New("invitation already accepted")
	ErrInvitationExpired         = errors.New("invitation expired")
	ErrInvitationInvalidPassword = errors.New("invalid invitation password")
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
	projectRepo    port.ProjectRepository
	memberRepo     port.ProjectMemberRepository
	userRepo       port.UserRepository
	noteRepo       port.NoteRepository
	diagramRepo    port.DiagramRepository
	invitationRepo port.InvitationRepository
	argon2Params   *Argon2Params
}

func NewProjectService(
	projectRepo port.ProjectRepository,
	memberRepo port.ProjectMemberRepository,
	userRepo port.UserRepository,
	noteRepo port.NoteRepository,
	diagramRepo port.DiagramRepository,
	invitationRepo port.InvitationRepository,
	argon2Params *Argon2Params,
) *ProjectService {
	return &ProjectService{
		projectRepo:    projectRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		noteRepo:       noteRepo,
		diagramRepo:    diagramRepo,
		invitationRepo: invitationRepo,
		argon2Params:   argon2Params,
	}
}

// CreateProject creates a new project with the creator as owner
func (s *ProjectService) CreateProject(
	ctx context.Context,
	userID primitive.ObjectID,
	name, description string,
	secretPassphrase string,
	secretSigningPrivateKey, signingPublicKey string,
	userPublicKey string, userEncryptedPrivateKey string,
) (*domain.Project, error) {
	project := &domain.Project{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: description,
		KeyEpoch:    "0",
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	// Add creator as owner
	member := &domain.ProjectMember{
		ProjectID:           project.ID,
		UserID:              userID,
		Role:                "owner",
		Permissions:         RolePresets["owner"],
		PublicKey:           userPublicKey,
		EncryptedPrivateKey: userEncryptedPrivateKey,
		Keyrings: []domain.ProjectMemberKeyring{
			{
				Epoch:                   "0",
				SecretPassphrase:        secretPassphrase,
				SecretSigningPrivateKey: secretSigningPrivateKey,
				SigningPublicKey:        signingPublicKey,
			},
		},
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

// CreateInvitation creates a new project invitation
func (s *ProjectService) CreateInvitation(
	ctx context.Context,
	projectID, inviterUserID, inviteeUserID primitive.ObjectID,
	role string,
	permissions []string,
	encryptedKeyrings string,
) (*domain.Invitation, error) {
	// Check permission
	if err := s.HasPermission(ctx, projectID, inviterUserID, domain.PermissionManageProject); err != nil {
		return nil, err
	}

	// Fetch project to get current KeyEpoch
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Check for existing pending invitation for this user in this project
	// and mark it as expired to prevent duplicates but keep history
	if !inviteeUserID.IsZero() {
		existingInv, err := s.invitationRepo.FindByProjectAndInvitee(ctx, projectID, inviteeUserID)
		if err == nil && existingInv != nil {
			// Found existing, mark as expired
			existingInv.Status = domain.InvitationStatusExpired
			_ = s.invitationRepo.Update(ctx, existingInv)
		}
	}

	invitation := &domain.Invitation{
		ID:                primitive.NewObjectID(),
		ProjectID:         projectID,
		InviterUserID:     inviterUserID,
		InviteeUserID:     inviteeUserID,
		Role:              role,
		Permissions:       permissions,
		EncryptedKeyrings: encryptedKeyrings,
		KeyEpoch:          project.KeyEpoch,
		Status:            domain.InvitationStatusPending,
	}

	result, err := s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetInvitation fetches an invitation by ID
func (s *ProjectService) GetInvitation(
	ctx context.Context,
	invitationID primitive.ObjectID,
) (*domain.Invitation, error) {
	invitation, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	return invitation, nil
}

// AcceptInvitation accepts an invitation and creates a project member
func (s *ProjectService) AcceptInvitation(
	ctx context.Context,
	invitationID, acceptingUserID primitive.ObjectID,
	keyrings []domain.ProjectMemberKeyring,
	publicKey, encryptedPrivateKey string,
) (primitive.ObjectID, error) {
	// Fetch invitation
	invitation, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return primitive.NilObjectID, ErrInvitationNotFound
		}
		return primitive.NilObjectID, err
	}

	// Check status
	if invitation.Status == domain.InvitationStatusAccepted {
		return primitive.NilObjectID, ErrInvitationAlreadyAccepted
	}
	if invitation.Status == domain.InvitationStatusExpired {
		return primitive.NilObjectID, ErrInvitationExpired
	}

	// Fetch project to check KeyEpoch
	project, err := s.projectRepo.FindByID(ctx, invitation.ProjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return primitive.NilObjectID, ErrProjectNotFound
		}
		return primitive.NilObjectID, err
	}

	// Verify KeyEpoch matches
	if project.KeyEpoch != invitation.KeyEpoch {
		// Mark as expired
		invitation.Status = domain.InvitationStatusExpired
		_ = s.invitationRepo.Update(ctx, invitation)
		return primitive.NilObjectID, ErrInvitationExpired
	}

	// Check if user is already a member
	existingMember, err := s.memberRepo.FindByProjectAndUser(ctx, invitation.ProjectID, acceptingUserID)
	if err == nil && existingMember != nil {
		// User is already a member. Check if this is a key rotation (new epoch)
		// Check if member already has keyring for this epoch
		hasEpoch := false
		for _, k := range existingMember.Keyrings {
			if k.Epoch == invitation.KeyEpoch {
				hasEpoch = true
				break
			}
		}

		if hasEpoch {
			return primitive.NilObjectID, ErrMemberAlreadyExists
		}

		// Update member with new keyrings
		existingMember.Keyrings = append(existingMember.Keyrings, keyrings...)
		if err := s.memberRepo.Update(ctx, existingMember); err != nil {
			return primitive.NilObjectID, err
		}

		// Mark invitation as accepted
		invitation.Status = domain.InvitationStatusAccepted
		if err := s.invitationRepo.Update(ctx, invitation); err != nil {
			return invitation.ProjectID, nil
		}

		return invitation.ProjectID, nil
	}

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return primitive.NilObjectID, err
	}

	// Create project member
	member := &domain.ProjectMember{
		ProjectID:           invitation.ProjectID,
		UserID:              acceptingUserID,
		Role:                invitation.Role,
		Permissions:         invitation.Permissions,
		Keyrings:            keyrings,
		PublicKey:           publicKey,
		EncryptedPrivateKey: encryptedPrivateKey,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return primitive.NilObjectID, err
	}

	// Mark invitation as accepted
	invitation.Status = domain.InvitationStatusAccepted
	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		// Non-critical: member was already created
		return invitation.ProjectID, nil
	}

	// Cleanup: Mark any other pending invitations for this user in this project as expired
	// This handles cases where multiple invitations might have been created (rare but possible safely)
	// or simply cleans up state.
	if !acceptingUserID.IsZero() {
		otherInv, err := s.invitationRepo.FindByProjectAndInvitee(ctx, invitation.ProjectID, acceptingUserID)
		if err == nil && otherInv != nil && otherInv.ID != invitation.ID {
			otherInv.Status = domain.InvitationStatusExpired
			_ = s.invitationRepo.Update(ctx, otherInv)
		}
	}

	return invitation.ProjectID, nil
}

// GetProjectInvitations lists invitations for a project
func (s *ProjectService) GetProjectInvitations(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	offset, limit int,
) ([]*domain.Invitation, int64, error) {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return nil, 0, err
	}

	return s.invitationRepo.FindByProjectID(ctx, projectID, offset, limit)
}

// GetUserInvitations lists invitations for the current user
func (s *ProjectService) GetUserInvitations(
	ctx context.Context,
	userID primitive.ObjectID,
	offset, limit int,
) ([]*domain.Invitation, int64, error) {
	return s.invitationRepo.FindByInviteeID(ctx, userID, offset, limit)
}

// RevokeInvitation revokes a pending invitation
func (s *ProjectService) RevokeInvitation(
	ctx context.Context,
	projectID, userID, invitationID primitive.ObjectID,
) error {
	// Check permission
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	// Verify invitation exists and belongs to this project
	invitation, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrInvitationNotFound
		}
		return err
	}

	if invitation.ProjectID != projectID {
		return ErrInvitationNotFound
	}

	if invitation.Status != domain.InvitationStatusPending {
		return ErrInvitationAlreadyAccepted
	}

	return s.invitationRepo.Delete(ctx, invitationID)
}

// RotateProjectKeys updates the project key epoch and adds new keyrings for members
func (s *ProjectService) RotateProjectKeys(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	newKeyEpoch string,
	updates []domain.MemberKeyringUpdate,
) error {
	// Check permission (Owner only for security critical operations)
	if err := s.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return err
	}

	// 1. Update Project Epoch
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	project.KeyEpoch = newKeyEpoch
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	// 2. Update Members
	// We do this in a loop. ideally this should be a transaction but for now separate updates are okay
	// as long as the project epoch is updated first.
	// If a member update fails, they just won't be able to access new data until re-invited/fixed,
	// but security is maintained because the project epoch has changed.
	for _, update := range updates {
		memberUserID, err := primitive.ObjectIDFromHex(update.UserID)
		if err != nil {
			continue // Skip invalid user IDs
		}

		member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, memberUserID)
		if err != nil {
			continue // Member might have been removed concurrently, skip
		}

		// Append new keyring
		newKeyring := domain.ProjectMemberKeyring{
			Epoch:                   newKeyEpoch,
			SecretPassphrase:        update.EncryptedPassphrase,
			SecretSigningPrivateKey: update.EncryptedSigningKey,
			SigningPublicKey:        update.SigningPublicKey,
		}
		member.Keyrings = append(member.Keyrings, newKeyring)

		if err := s.memberRepo.Update(ctx, member); err != nil {
			logger.Error().Err(err).Str("project_id", projectID.Hex()).Str("user_id", update.UserID).Msg("Failed to update member keyring")
		} else {
			logger.Info().Str("project_id", projectID.Hex()).Str("user_id", update.UserID).Msg("Updated member keyring")
		}
	}

	return nil
}
