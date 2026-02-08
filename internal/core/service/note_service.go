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
func (s *NoteService) CreateNote(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	fileName, fileType string,
	encryptedContent *string,
	signature string,
) (*domain.Note, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionEditNote); err != nil {
		return nil, err
	}

	note := &domain.Note{
		ID:                        primitive.NewObjectID(),
		ProjectID:                 projectID,
		FileName:                  fileName,
		FileType:                  fileType,
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

// ListNotes retrieves all notes for a project with pagination
func (s *NoteService) ListNotes(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	offset, limit int,
) ([]*domain.Note, int64, error) {
	// Check permission
	if err := s.hasPermission(ctx, projectID, userID, domain.PermissionViewNote); err != nil {
		return nil, 0, err
	}

	return s.noteRepo.FindByProjectID(ctx, projectID, offset, limit)
}

// UpdateNote updates an existing note
func (s *NoteService) UpdateNote(
	ctx context.Context,
	noteID, userID primitive.ObjectID,
	fileName, fileType *string,
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
	if fileType != nil {
		note.FileType = *fileType
	}
	if encryptedContent != nil {
		note.EncryptedContent = encryptedContent
	}
	if signature != nil {
		note.EncryptedContentSignature = *signature
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
