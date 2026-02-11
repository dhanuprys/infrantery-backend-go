package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidID           = errors.New("invalid id format")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrResourceNotFound    = errors.New("resource not found")
)

type BreadcrumbService struct {
	projectRepo   port.ProjectRepository
	noteRepo      port.NoteRepository
	diagramRepo   port.DiagramRepository
	nodeRepo      port.NodeRepository
	nodeVaultRepo port.NodeVaultRepository
}

func NewBreadcrumbService(
	projectRepo port.ProjectRepository,
	noteRepo port.NoteRepository,
	diagramRepo port.DiagramRepository,
	nodeRepo port.NodeRepository,
	nodeVaultRepo port.NodeVaultRepository,
) *BreadcrumbService {
	return &BreadcrumbService{
		projectRepo:   projectRepo,
		noteRepo:      noteRepo,
		diagramRepo:   diagramRepo,
		nodeRepo:      nodeRepo,
		nodeVaultRepo: nodeVaultRepo,
	}
}

func (s *BreadcrumbService) GetBreadcrumbs(ctx context.Context, projectIDStr, resourceType, resourceIDStr string) (*dto.BreadcrumbResponse, error) {
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		return nil, ErrInvalidID
	}

	// Verify project exists
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	path := []dto.BreadcrumbItem{
		{
			Type:   "project",
			ID:     project.ID.Hex(),
			Label:  project.Name,
			Active: resourceType == "project",
		},
	}

	if resourceType == "project" {
		return &dto.BreadcrumbResponse{
			ProjectID: projectIDStr,
			Path:      path,
		}, nil
	}

	// Handle list views or strict ID parsing
	if resourceIDStr == "" {
		switch resourceType {
		case "note":
			path = append(path, dto.BreadcrumbItem{
				Type:   "note",
				Label:  "Notes",
				Active: true,
			})
			return &dto.BreadcrumbResponse{
				ProjectID: projectIDStr,
				Path:      path,
			}, nil
		case "vault":
			path = append(path, dto.BreadcrumbItem{
				Type:   "vault",
				Label:  "Vault",
				Active: true,
			})
			return &dto.BreadcrumbResponse{
				ProjectID: projectIDStr,
				Path:      path,
			}, nil
		}
	}

	resourceID, err := primitive.ObjectIDFromHex(resourceIDStr)
	if err != nil {
		return nil, ErrInvalidID
	}

	switch resourceType {
	case "note":
		return s.handleNoteBreadcrumb(ctx, projectID, resourceID, path)
	case "diagram":
		return s.handleDiagramBreadcrumb(ctx, projectID, resourceID, path)
	case "node":
		return s.handleNodeBreadcrumb(ctx, projectID, resourceID, path)
	case "vault":
		return s.handleVaultBreadcrumb(ctx, projectID, resourceID, path)
	case "node_vault":
		return s.handleNodeVaultListBreadcrumb(ctx, projectID, resourceID, path)
	default:
		return nil, ErrInvalidResourceType
	}
}

func (s *BreadcrumbService) handleNoteBreadcrumb(ctx context.Context, projectID, noteID primitive.ObjectID, basePath []dto.BreadcrumbItem) (*dto.BreadcrumbResponse, error) {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return nil, err
	}
	if note == nil || note.ProjectID != projectID {
		logger.Error().Msgf("Note not found or project mismatch: NoteID=%s, ProjectID=%s", noteID.Hex(), projectID.Hex())
		return nil, ErrResourceNotFound
	}

	// Fetch siblings (other notes in project) - fetch up to 100 for now
	notes, err := s.noteRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	siblings := make([]dto.BreadcrumbItem, 0, len(notes))
	for _, n := range notes {
		if n.ID != note.ID {
			siblings = append(siblings, dto.BreadcrumbItem{
				Type:  "note",
				ID:    n.ID.Hex(),
				Label: n.FileName,
			})
		}
	}

	path := append(basePath, dto.BreadcrumbItem{
		Type:     "note",
		ID:       note.ID.Hex(),
		Label:    note.FileName,
		Active:   true,
		Siblings: siblings,
	})

	return &dto.BreadcrumbResponse{
		ProjectID: projectID.Hex(),
		Path:      path,
	}, nil
}

func (s *BreadcrumbService) handleDiagramBreadcrumb(ctx context.Context, projectID, diagramID primitive.ObjectID, basePath []dto.BreadcrumbItem) (*dto.BreadcrumbResponse, error) {
	diagramPath, err := s.buildDiagramPath(ctx, projectID, diagramID)
	if err != nil {
		return nil, err
	}

	path := append(basePath, diagramPath...)
	// Mark last item as active
	path[len(path)-1].Active = true

	return &dto.BreadcrumbResponse{
		ProjectID: projectID.Hex(),
		Path:      path,
	}, nil
}

type diagramNode struct {
	diagram  *domain.Diagram
	siblings []dto.BreadcrumbItem
}

func (s *BreadcrumbService) buildDiagramPath(ctx context.Context, projectID, diagramID primitive.ObjectID) ([]dto.BreadcrumbItem, error) {
	var path []dto.BreadcrumbItem

	// We need to traverse up from the current diagram to the root (where ParentDiagramID is nil)
	// Because MongoDB doesn't support recursive graph queries easily without $graphLookup (which mgod might abstract)
	// We'll do iterative lookups up the chain. Given standard depth isn't huge, this is acceptable.

	currentID := &diagramID
	var chain []*diagramNode

	for currentID != nil {
		diagram, err := s.diagramRepo.FindByID(ctx, *currentID)
		if err != nil {
			return nil, err
		}
		if diagram == nil || diagram.ProjectID != projectID {
			logger.Error().Msgf("Diagram not found or project mismatch: DiagramID=%s, ProjectID=%s", currentID.Hex(), projectID.Hex())
			return nil, fmt.Errorf("diagram not found or project mismatch (ID: %s): %w", currentID.Hex(), ErrResourceNotFound)
		}

		// Fetch siblings for this level
		siblings, err := s.getDiagramSiblings(ctx, projectID, diagram.ParentDiagramID, diagram.ID)
		if err != nil {
			return nil, err
		}

		chain = append([]*diagramNode{{diagram: diagram, siblings: siblings}}, chain...) // Prepend
		currentID = diagram.ParentDiagramID
	}

	for _, node := range chain {
		path = append(path, dto.BreadcrumbItem{
			Type:     "diagram",
			ID:       node.diagram.ID.Hex(),
			Label:    node.diagram.DiagramName,
			Siblings: node.siblings,
		})
	}

	return path, nil
}

func (s *BreadcrumbService) getDiagramSiblings(ctx context.Context, projectID primitive.ObjectID, parentID *primitive.ObjectID, excludeID primitive.ObjectID) ([]dto.BreadcrumbItem, error) {
	// This requires a new method on DiagramRepo to find by parent or root
	// Since we might not have it, we'll fetch all project diagrams and filter in memory for now,
	// OR we assume existing method FindByProjectID exists.
	// Optimization: Add FindByParentID to repo later.
	// Fetching 100 diagrams for now
	allDiagrams, _, err := s.diagramRepo.FindByProjectID(ctx, projectID, false, 0, 100)
	if err != nil {
		return nil, err
	}

	var siblings []dto.BreadcrumbItem
	for _, d := range allDiagrams {
		// Check if sibling (same parent)
		isSibling := false
		if parentID == nil {
			if d.ParentDiagramID == nil {
				isSibling = true
			}
		} else {
			if d.ParentDiagramID != nil && *d.ParentDiagramID == *parentID {
				isSibling = true
			}
		}

		if isSibling && d.ID != excludeID {
			siblings = append(siblings, dto.BreadcrumbItem{
				Type:  "diagram",
				ID:    d.ID.Hex(),
				Label: d.DiagramName,
			})
		}
	}
	return siblings, nil
}

func (s *BreadcrumbService) handleNodeBreadcrumb(ctx context.Context, projectID, nodeID primitive.ObjectID, basePath []dto.BreadcrumbItem) (*dto.BreadcrumbResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrResourceNotFound
	}

	// Build diagram path
	diagramPath, err := s.buildDiagramPath(ctx, projectID, node.DiagramID)
	if err != nil {
		return nil, err
	}

	path := append(basePath, diagramPath...)
	path = append(path, dto.BreadcrumbItem{
		Type:   "node",
		ID:     node.ID.Hex(),
		Label:  "Node", // Nodes might not have names in current model, using generic label
		Active: true,
	})

	return &dto.BreadcrumbResponse{
		ProjectID: projectID.Hex(),
		Path:      path,
	}, nil
}

func (s *BreadcrumbService) handleVaultBreadcrumb(ctx context.Context, projectID, vaultID primitive.ObjectID, basePath []dto.BreadcrumbItem) (*dto.BreadcrumbResponse, error) {
	vault, err := s.nodeVaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	if vault == nil {
		return nil, ErrResourceNotFound
	}

	// Fetch Node
	node, err := s.nodeRepo.FindByID(ctx, vault.NodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrResourceNotFound
	}

	// Build diagram path
	diagramPath, err := s.buildDiagramPath(ctx, projectID, node.DiagramID)
	if err != nil {
		return nil, err
	}

	path := append(basePath, diagramPath...)
	path = append(path, dto.BreadcrumbItem{
		Type:  "node",
		ID:    node.ID.Hex(),
		Label: "Node",
	})
	path = append(path, dto.BreadcrumbItem{
		Type:   "vault",
		ID:     vault.ID.Hex(),
		Label:  fmt.Sprintf("Vault (%s)", vault.Type),
		Active: true,
	})

	return &dto.BreadcrumbResponse{
		ProjectID: projectID.Hex(),
		Path:      path,
	}, nil
}

func (s *BreadcrumbService) handleNodeVaultListBreadcrumb(ctx context.Context, projectID, nodeID primitive.ObjectID, basePath []dto.BreadcrumbItem) (*dto.BreadcrumbResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		logger.Error().Msgf("Node not found for breadcrumb: NodeID=%s", nodeID.Hex())
		return nil, fmt.Errorf("node not found (ID: %s): %w", nodeID.Hex(), ErrResourceNotFound)
	}

	// Build diagram path
	diagramPath, err := s.buildDiagramPath(ctx, projectID, node.DiagramID)
	if err != nil {
		return nil, err
	}

	path := append(basePath, diagramPath...)
	path = append(path, dto.BreadcrumbItem{
		Type:  "node",
		ID:    node.ID.Hex(),
		Label: "Node",
	})
	path = append(path, dto.BreadcrumbItem{
		Type:   "node_vault",
		ID:     node.ID.Hex(), // Use node ID as we are listing vault for this node
		Label:  "Vault",
		Active: true,
	})

	return &dto.BreadcrumbResponse{
		ProjectID: projectID.Hex(),
		Path:      path,
	}, nil
}
