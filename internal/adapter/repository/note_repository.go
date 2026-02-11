package repository

import (
	"context"

	"github.com/Lyearn/mgod"
	"github.com/Lyearn/mgod/schema/schemaopt"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *noteRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Note, error) {
	filter := bson.M{"project_id": projectID}

	// Sort alphabetically by file name
	opts := options.Find().SetSort(bson.D{{Key: "file_name", Value: 1}}).SetCollation(&options.Collation{Locale: "en", Strength: 1})

	// Get all notes
	allNotes, err := r.model.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	// Convert to pointers
	result := make([]*domain.Note, 0, len(allNotes))
	for i := range allNotes {
		result = append(result, &allNotes[i])
	}

	return result, nil
}

func (r *noteRepository) Update(ctx context.Context, note *domain.Note) error {
	filter := bson.M{"_id": note.ID}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "file_name", Value: note.FileName},
			{Key: "type", Value: note.Type},
			{Key: "parent_id", Value: note.ParentID},
			{Key: "icon", Value: note.Icon},
			{Key: "encrypted_content", Value: note.EncryptedContent},
			{Key: "encrypted_content_signature", Value: note.EncryptedContentSignature},
		}},
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
