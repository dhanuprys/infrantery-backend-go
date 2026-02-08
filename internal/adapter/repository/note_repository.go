package repository

import (
	"context"

	"github.com/Lyearn/mgod"
	"github.com/Lyearn/mgod/schema/schemaopt"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type noteRepository struct {
	model mgod.EntityMongoModel[domain.Note]
}

func NewNoteRepository(collectionName string) (port.NoteRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.Note{}, opts)
	if err != nil {
		return nil, err
	}

	return &noteRepository{model: model}, nil
}

func (r *noteRepository) Create(ctx context.Context, note *domain.Note) error {
	result, err := r.model.InsertOne(ctx, *note)
	if err != nil {
		return err
	}
	note.ID = result.ID
	return nil
}

func (r *noteRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Note, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *noteRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID, offset, limit int) ([]*domain.Note, int64, error) {
	filter := bson.M{"project_id": projectID}

	// Get total count
	allNotes, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allNotes))

	// Apply pagination
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allNotes) {
		return []*domain.Note{}, totalCount, nil
	}
	if endIdx > len(allNotes) {
		endIdx = len(allNotes)
	}
	paginatedNotes := allNotes[startIdx:endIdx]

	// Convert to pointers
	result := make([]*domain.Note, 0, len(paginatedNotes))
	for i := range paginatedNotes {
		result = append(result, &paginatedNotes[i])
	}

	return result, totalCount, nil
}

func (r *noteRepository) Update(ctx context.Context, note *domain.Note) error {
	filter := bson.M{"_id": note.ID}
	update := bson.M{
		"$set": bson.M{
			"file_name":                   note.FileName,
			"file_type":                   note.FileType,
			"encrypted_content":           note.EncryptedContent,
			"encrypted_content_signature": note.EncryptedContentSignature,
		},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *noteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}

func (r *noteRepository) DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}
