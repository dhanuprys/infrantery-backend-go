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
	ErrNoteNotFound     = errors.New("note not found")
	ErrNoteAccessDenied = errors.New("note access denied")
)

type NoteService struct {
	noteRepo    port.NoteRepository
	memberRepo  port.ProjectMemberRepository
	projectRepo port.ProjectRepository
}

func NewNoteService(
	noteRepo port.NoteRepository,
	memberRepo port.ProjectMemberRepository,
	projectRepo port.ProjectRepository,
) *NoteService {
	return &NoteService{
		noteRepo:    noteRepo,
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
	}
}

// CreateNote creates a new note in a project
// CreateNote creates a new note in a project
func (s *NoteService) CreateNote(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	parentID *primitive.ObjectID,
	noteType string,
	fileName string,
	icon string,
	encryptedContent *string,
	signature *string,
) (*domain.Note, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionEditNote); err != nil {
		return nil, err
	}

	// Verify parent if provided
	if parentID != nil {
		if err := s.verifyParent(ctx, *parentID, projectID); err != nil {
			return nil, err
		}
	}

	note := &domain.Note{
		ID:                        primitive.NewObjectID(),
		ProjectID:                 projectID,
		ParentID:                  parentID,
		Type:                      noteType,
		FileName:                  fileName,
		Icon:                      icon,
		EncryptedContent:          encryptedContent,
		EncryptedContentSignature: signature,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

// GetNote retrieves a specific note
func (s *NoteService) GetNote(
	ctx context.Context,
	noteID, userID primitive.ObjectID,
) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}

	// Check permission
	if err := s.hasPermission(ctx, note.ProjectID, userID, domain.PermissionViewNote); err != nil {
		return nil, err
	}

	return note, nil
}

// ListNotes retrieves all notes for a project
func (s *NoteService) ListNotes(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
) ([]*domain.Note, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionViewNote); err != nil {
		return nil, err
	}

	// Fetch all notes (no pagination)
	return s.noteRepo.FindByProjectID(ctx, projectID)
}

// UpdateNote updates an existing note
func (s *NoteService) UpdateNote(
	ctx context.Context,
	noteID, userID primitive.ObjectID,
	fileName *string,
	parentID *string, // Receive as string pointer to distinguish unset vs empty (though usually ObjectID)
	icon *string,
	encryptedContent, signature *string,
) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}

	// Check permission
	if err := s.hasPermission(ctx, note.ProjectID, userID, domain.PermissionEditNote); err != nil {
		return nil, err
	}

	// Update fields if provided
	if fileName != nil {
		note.FileName = *fileName
	}
	if parentID != nil {
		if *parentID == "" {
			note.ParentID = nil
		} else {
			pid, err := primitive.ObjectIDFromHex(*parentID)
			if err == nil {
				// Verify new parent
				if err := s.verifyParent(ctx, pid, note.ProjectID); err != nil {
					return nil, err
				}
				note.ParentID = &pid
			}
		}
	}
	if icon != nil {
		note.Icon = *icon
	}
	if encryptedContent != nil {
		note.EncryptedContent = encryptedContent
	}
	if signature != nil {
		note.EncryptedContentSignature = signature
	}

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

// DeleteNote deletes a note
func (s *NoteService) DeleteNote(
	ctx context.Context,
	noteID, userID primitive.ObjectID,
) error {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoteNotFound
		}
		return err
	}

	// Check permission
	if err := s.hasPermission(ctx, note.ProjectID, userID, domain.PermissionEditNote); err != nil {
		return err
	}

	return s.noteRepo.Delete(ctx, noteID)
}

// verifyParent checks if the parent ID exists and is a folder
func (s *NoteService) verifyParent(ctx context.Context, parentID primitive.ObjectID, projectID primitive.ObjectID) error {
	parent, err := s.noteRepo.FindByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("parent folder not found")
		}
		return err
	}

	if parent.ProjectID != projectID {
		return errors.New("parent folder belongs to a different project")
	}

	if parent.Type != "folder" {
		return errors.New("parent is not a folder")
	}

	return nil
}

// hasPermission checks if user has a specific permission for the project
func (s *NoteService) hasPermission(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	permission string,
) error {
	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoteAccessDenied
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
