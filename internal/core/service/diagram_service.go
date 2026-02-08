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
	ErrDiagramNotFound     = errors.New("diagram not found")
	ErrDiagramAccessDenied = errors.New("diagram access denied")
)

type DiagramService struct {
	diagramRepo port.DiagramRepository
	memberRepo  port.ProjectMemberRepository
	projectRepo port.ProjectRepository
	nodeRepo    port.NodeRepository
}

func NewDiagramService(
	diagramRepo port.DiagramRepository,
	memberRepo port.ProjectMemberRepository,
	projectRepo port.ProjectRepository,
	nodeRepo port.NodeRepository,
) *DiagramService {
	return &DiagramService{
		diagramRepo: diagramRepo,
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
		nodeRepo:    nodeRepo,
	}
}

// CreateDiagram creates a new diagram in a project
func (s *DiagramService) CreateDiagram(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	diagramName, description string,
	parentDiagramID *primitive.ObjectID,
	encryptedData *string,
	signature string,
) (*domain.Diagram, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionEditDiagram); err != nil {
		return nil, err
	}

	diagram := &domain.Diagram{
		ID:                     primitive.NewObjectID(),
		ProjectID:              projectID,
		DiagramName:            diagramName,
		Description:            description,
		ParentDiagramID:        parentDiagramID,
		EncryptedData:          encryptedData,
		EncryptedDataSignature: signature,
	}

	if err := s.diagramRepo.Create(ctx, diagram); err != nil {
		return nil, err
	}

	return diagram, nil
}

// GetDiagram retrieves a specific diagram
func (s *DiagramService) GetDiagram(
	ctx context.Context,
	diagramID, userID primitive.ObjectID,
) (*domain.Diagram, error) {
	diagram, err := s.diagramRepo.FindByID(ctx, diagramID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDiagramNotFound
		}
		return nil, err
	}

	// Check permission
	if err := s.hasPermission(ctx, diagram.ProjectID, userID, domain.PermissionViewDiagram); err != nil {
		return nil, err
	}

	return diagram, nil
}

// ListDiagrams retrieves all diagrams for a project with pagination
func (s *DiagramService) ListDiagrams(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	offset, limit int,
) ([]*domain.Diagram, int64, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionViewDiagram); err != nil {
		return nil, 0, err
	}

	return s.diagramRepo.FindByProjectID(ctx, projectID, offset, limit)
}

// UpdateDiagram updates an existing diagram
func (s *DiagramService) UpdateDiagram(
	ctx context.Context,
	diagramID, userID primitive.ObjectID,
	diagramName, description *string,
	encryptedData, signature *string,
) (*domain.Diagram, error) {
	diagram, err := s.diagramRepo.FindByID(ctx, diagramID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDiagramNotFound
		}
		return nil, err
	}

	// Check permission
	if err := s.hasPermission(ctx, diagram.ProjectID, userID, domain.PermissionEditDiagram); err != nil {
		return nil, err
	}

	// Update fields if provided
	if diagramName != nil {
		diagram.DiagramName = *diagramName
	}
	if description != nil {
		diagram.Description = *description
	}
	if encryptedData != nil {
		diagram.EncryptedData = encryptedData
	}
	if signature != nil {
		diagram.EncryptedDataSignature = *signature
	}

	if err := s.diagramRepo.Update(ctx, diagram); err != nil {
		return nil, err
	}

	return diagram, nil
}

// DeleteDiagram deletes a diagram
func (s *DiagramService) DeleteDiagram(
	ctx context.Context,
	diagramID, userID primitive.ObjectID,
) error {
	diagram, err := s.diagramRepo.FindByID(ctx, diagramID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrDiagramNotFound
		}
		return err
	}

	// Check permission
	if err := s.hasPermission(ctx, diagram.ProjectID, userID, domain.PermissionEditDiagram); err != nil {
		return err
	}

	// Delete all nodes associated with this diagram
	if err := s.nodeRepo.DeleteByDiagramID(ctx, diagramID); err != nil {
		return err
	}

	return s.diagramRepo.Delete(ctx, diagramID)
}

// hasPermission checks if user has a specific permission for the project
func (s *DiagramService) hasPermission(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	permission string,
) error {
	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrDiagramAccessDenied
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
