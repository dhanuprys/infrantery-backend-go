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

type diagramRepository struct {
	model mgod.EntityMongoModel[domain.Diagram]
}

func NewDiagramRepository(collectionName string) (port.DiagramRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.Diagram{}, opts)
	if err != nil {
		return nil, err
	}

	return &diagramRepository{model: model}, nil
}

func (r *diagramRepository) Create(ctx context.Context, diagram *domain.Diagram) error {
	result, err := r.model.InsertOne(ctx, *diagram)
	if err != nil {
		return err
	}
	diagram.ID = result.ID
	return nil
}

func (r *diagramRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Diagram, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *diagramRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID, rootOnly bool, offset, limit int) ([]*domain.Diagram, int64, error) {
	filter := bson.M{"project_id": projectID}
	if rootOnly {
		filter["parent_diagram_id"] = nil
	}

	// Get total count
	allDiagrams, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allDiagrams))

	// Apply pagination
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allDiagrams) {
		return []*domain.Diagram{}, totalCount, nil
	}
	if endIdx > len(allDiagrams) {
		endIdx = len(allDiagrams)
	}
	paginatedDiagrams := allDiagrams[startIdx:endIdx]

	// Convert to pointers
	result := make([]*domain.Diagram, 0, len(paginatedDiagrams))
	for i := range paginatedDiagrams {
		result = append(result, &paginatedDiagrams[i])
	}

	return result, totalCount, nil
}

func (r *diagramRepository) FindAllByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Diagram, error) {
	diagrams, err := r.model.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Diagram, 0, len(diagrams))
	for i := range diagrams {
		result = append(result, &diagrams[i])
	}
	return result, nil
}

func (r *diagramRepository) Update(ctx context.Context, diagram *domain.Diagram) error {
	filter := bson.M{"_id": diagram.ID}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "diagram_name", Value: diagram.DiagramName},
			{Key: "description", Value: diagram.Description},
			{Key: "parent_diagram_id", Value: diagram.ParentDiagramID},
			{Key: "encrypted_data", Value: diagram.EncryptedData},
			{Key: "encrypted_data_signature", Value: diagram.EncryptedDataSignature},
		}},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *diagramRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}

func (r *diagramRepository) DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}
