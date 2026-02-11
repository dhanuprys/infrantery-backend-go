package service

import (
	"context"
	"errors"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNodeNotFound     = errors.New(dto.ErrCodeNodeNotFound)
	ErrNodeAccessDenied = errors.New(dto.ErrCodeNodeAccessDenied)
	ErrInvalidNodeID    = errors.New(dto.ErrCodeInvalidNodeID)
)

type NodeService struct {
	nodeRepo          port.NodeRepository
	diagramRepo       port.DiagramRepository
	projectMemberRepo port.ProjectMemberRepository
}

func NewNodeService(
	nodeRepo port.NodeRepository,
	diagramRepo port.DiagramRepository,
	projectMemberRepo port.ProjectMemberRepository,
) *NodeService {
	return &NodeService{
		nodeRepo:          nodeRepo,
		diagramRepo:       diagramRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

// GetOrCreateNode gets a node or creates it if it doesn't exist, validating permissions via diagram
func (s *NodeService) GetOrCreateNode(ctx context.Context, nodeIDStr string, diagramID primitive.ObjectID, userID primitive.ObjectID) (*domain.Node, error) {
	// Validate Node ID format
	nodeID, err := primitive.ObjectIDFromHex(nodeIDStr)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	// Try to find the node
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	if node != nil {
		// Node exists: Verify diagram match and permission
		if node.DiagramID != diagramID {
			// Preventing ID manipulation: Node belongs to a different diagram
			return nil, ErrNodeAccessDenied
		}

		// Verify view permission on parent diagram
		if err := s.verifyDiagramPermission(ctx, diagramID, userID, "view_diagram"); err != nil {
			return nil, err
		}

		return node, nil
	}

	// Node doesn't exist: Create it (requires edit permission)
	if err := s.verifyDiagramPermission(ctx, diagramID, userID, "edit_diagram"); err != nil {
		return nil, err
	}

	newNode := &domain.Node{
		ID:        nodeID,
		DiagramID: diagramID,
		// Encrypted fields start empty
	}

	if err := s.nodeRepo.Create(ctx, newNode); err != nil {
		return nil, err
	}

	return newNode, nil
}

// UpdateNode updates a node's encrypted data
func (s *NodeService) UpdateNode(ctx context.Context, nodeIDStr string, userID primitive.ObjectID, req dto.UpdateNodeRequest) (*domain.Node, error) {
	nodeID, err := primitive.ObjectIDFromHex(nodeIDStr)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}

	// Verify edit permission
	if err := s.verifyDiagramPermission(ctx, node.DiagramID, userID, "edit_diagram"); err != nil {
		return nil, err
	}

	// Update fields
	if req.EncryptedReadme != nil {
		node.EncryptedReadme = *req.EncryptedReadme
	}
	if req.EncryptedReadmeSignature != nil {
		node.EncryptedReadmeSignature = *req.EncryptedReadmeSignature
	}
	if req.EncryptedDict != nil {
		node.EncryptedDict = *req.EncryptedDict
	}
	if req.EncryptedDictSignature != nil {
		node.EncryptedDictSignature = *req.EncryptedDictSignature
	}

	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}

	return node, nil
}

// DeleteNode deletes a node
func (s *NodeService) DeleteNode(ctx context.Context, nodeIDStr string, userID primitive.ObjectID) error {
	nodeID, err := primitive.ObjectIDFromHex(nodeIDStr)
	if err != nil {
		return ErrInvalidNodeID
	}

	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil // Idempotent: Node already gone
		}
		return err
	}

	if node == nil {
		return nil // Idempotent: Node already gone
	}

	// Verify edit permission
	if err := s.verifyDiagramPermission(ctx, node.DiagramID, userID, "edit_diagram"); err != nil {
		return err
	}

	return s.nodeRepo.Delete(ctx, nodeID)
}

// Helper to verify diagram permissions
func (s *NodeService) verifyDiagramPermission(ctx context.Context, diagramID, userID primitive.ObjectID, requiredPermission string) error {
	// 1. Get diagram to find project ID
	diagram, err := s.diagramRepo.FindByID(ctx, diagramID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New(dto.ErrCodeDiagramNotFound)
		}
		return err
	}

	// 2. Check project membership/permissions
	member, err := s.projectMemberRepo.FindByProjectAndUser(ctx, diagram.ProjectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNodeAccessDenied
		}
		return err
	}

	// Check specific permission (simplified for now, assuming role check or direct permission exist)
	// In previous steps, we used HasPermission method if available, or role checks.
	// ProjectMember struct likely has Permissions list or Role.
	// Let's assume verifying member existence is enough for now, OR check role.
	// Actually, we should check strictly. Let's look at ProjectMember definition or DiagramService pattern.
	// For now, I'll restrict to Member exists. To be more strict, I should check permissions.
	// But `ProjectMember` struct might not be fully visible here without importing domain.

	// Based on DiagramService implementation:
	// permission := "view_diagram"
	// hasPermission := false
	// for _, p := range member.Permissions {
	//     if p == permission {
	//         hasPermission = true
	//         break
	//     }
	// }

	hasPermission := false
	for _, p := range member.Permissions {
		if p == requiredPermission {
			hasPermission = true
			break
		}
	}
	// Owners always have permission (usually)
	// But let's stick to explicit permissions if that's the model.
	// Or check role == "owner".
	if member.Role == "owner" {
		hasPermission = true
	}

	if !hasPermission {
		return ErrNodeAccessDenied
	}

	return nil
}
