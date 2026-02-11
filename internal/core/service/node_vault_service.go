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
	ErrVaultItemNotFound = errors.New(dto.ErrCodeVaultItemNotFound)
	ErrVaultAccessDenied = errors.New(dto.ErrCodeVaultAccessDenied)
	ErrInvalidRequest    = errors.New(dto.ErrCodeInvalidRequest)
)

type NodeVaultService struct {
	nodeVaultRepo     port.NodeVaultRepository
	nodeRepo          port.NodeRepository
	diagramRepo       port.DiagramRepository
	projectMemberRepo port.ProjectMemberRepository
}

func NewNodeVaultService(
	nodeVaultRepo port.NodeVaultRepository,
	nodeRepo port.NodeRepository,
	diagramRepo port.DiagramRepository,
	projectMemberRepo port.ProjectMemberRepository,
) *NodeVaultService {
	return &NodeVaultService{
		nodeVaultRepo:     nodeVaultRepo,
		nodeRepo:          nodeRepo,
		diagramRepo:       diagramRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

// CreateVaultItem creates a new vault item for a node
func (s *NodeVaultService) CreateVaultItem(ctx context.Context, nodeIDStr string, projectID primitive.ObjectID, userID primitive.ObjectID, req dto.CreateNodeVaultRequest) (*domain.NodeVault, error) {
	nodeID, err := primitive.ObjectIDFromHex(nodeIDStr)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	// 1. Verify Edit Permission using passed ProjectID
	if err := s.verifyProjectPermission(ctx, projectID, userID, "edit_vault"); err != nil {
		return nil, err
	}

	vaultItem := &domain.NodeVault{
		NodeId:                  nodeID,
		ProjectId:               projectID,
		Label:                   req.Label,
		Type:                    req.Type,
		EncryptedValue:          &req.EncryptedValue,
		EncryptedValueSignature: &req.EncryptedValueSignature,
	}

	if err := s.nodeVaultRepo.Create(ctx, vaultItem); err != nil {
		return nil, err
	}

	return vaultItem, nil
}

// GetVaultItem gets a specific vault item by ID
func (s *NodeVaultService) GetVaultItem(ctx context.Context, vaultIDStr string, userID primitive.ObjectID) (*domain.NodeVault, error) {
	vaultID, err := primitive.ObjectIDFromHex(vaultIDStr)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	vaultItem, err := s.nodeVaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrVaultItemNotFound
		}
		return nil, err
	}

	// Verify Edit/View Permission (using view_vault as minimum)
	if err := s.verifyProjectPermission(ctx, vaultItem.ProjectId, userID, "view_vault"); err != nil {
		return nil, err
	}

	return vaultItem, nil
}

// ListVaultItems lists all vault items for a node
func (s *NodeVaultService) ListVaultItems(ctx context.Context, nodeIDStr string, projectID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.NodeVault, error) {
	nodeID, err := primitive.ObjectIDFromHex(nodeIDStr)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	// 1. Verify View Permission using passed ProjectID
	if err := s.verifyProjectPermission(ctx, projectID, userID, "view_vault"); err != nil {
		return nil, err
	}

	items, err := s.nodeVaultRepo.FindByNodeID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// For list view, we don't return encrypted values (lazy loading)
	for _, item := range items {
		item.EncryptedValue = nil
		item.EncryptedValueSignature = nil
	}

	return items, nil
}

// UpdateVaultItem updates a vault item
func (s *NodeVaultService) UpdateVaultItem(ctx context.Context, vaultIDStr string, userID primitive.ObjectID, req dto.UpdateNodeVaultRequest) (*domain.NodeVault, error) {
	vaultID, err := primitive.ObjectIDFromHex(vaultIDStr)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	vaultItem, err := s.nodeVaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrVaultItemNotFound
		}
		return nil, err
	}

	// Verify Edit Permission using denormalized ProjectID
	if err := s.verifyProjectPermission(ctx, vaultItem.ProjectId, userID, "edit_vault"); err != nil {
		return nil, err
	}

	if req.Label != nil {
		vaultItem.Label = *req.Label
	}
	if req.EncryptedValue != nil {
		vaultItem.EncryptedValue = req.EncryptedValue
	}
	if req.EncryptedValueSignature != nil {
		vaultItem.EncryptedValueSignature = req.EncryptedValueSignature
	}

	if err := s.nodeVaultRepo.Update(ctx, vaultItem); err != nil {
		return nil, err
	}

	return vaultItem, nil
}

// DeleteVaultItem deletes a vault item
func (s *NodeVaultService) DeleteVaultItem(ctx context.Context, vaultIDStr string, userID primitive.ObjectID) error {
	vaultID, err := primitive.ObjectIDFromHex(vaultIDStr)
	if err != nil {
		return ErrInvalidRequest
	}

	vaultItem, err := s.nodeVaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrVaultItemNotFound
		}
		return err
	}

	// Verify Edit Permission using denormalized ProjectID
	if err := s.verifyProjectPermission(ctx, vaultItem.ProjectId, userID, "edit_vault"); err != nil {
		return err
	}

	return s.nodeVaultRepo.Delete(ctx, vaultID)
}

func (s *NodeVaultService) verifyProjectPermission(ctx context.Context, projectID, userID primitive.ObjectID, permission string) error {
	member, err := s.projectMemberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrVaultAccessDenied
		}
		return err
	}

	// Check permission or owner role
	if member.Role == "owner" {
		return nil
	}
	for _, p := range member.Permissions {
		if p == permission {
			return nil
		}
	}

	return ErrVaultAccessDenied // Or ErrInsufficientPermission
}
