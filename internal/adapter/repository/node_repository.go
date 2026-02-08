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

type nodeRepository struct {
	model mgod.EntityMongoModel[domain.Node]
}

func NewNodeRepository(collectionName string) (port.NodeRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.Node{}, opts)
	if err != nil {
		return nil, err
	}

	return &nodeRepository{model: model}, nil
}

func (r *nodeRepository) Create(ctx context.Context, node *domain.Node) error {
	// ID is already set by caller (frontend generated), mgod will respect it if _id field is set
	// However, usually mgod inserts and returns the ID.
	// Since we defined bson:"_id" on ID field, mgod should handle it.
	// If ID is zero, mgod/mongo driver will generate one.
	// We want to force the ID provided in the struct if it's set.
	// The mongo driver supports this.
	_, err := r.model.InsertOne(ctx, *node)
	return err
}

func (r *nodeRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Node, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *nodeRepository) FindByDiagramID(ctx context.Context, diagramID primitive.ObjectID, offset, limit int) ([]*domain.Node, int64, error) {
	filter := bson.M{"diagram_id": diagramID}

	// Get total count
	allNodes, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allNodes))

	// Apply pagination
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allNodes) {
		return []*domain.Node{}, totalCount, nil
	}
	if endIdx > len(allNodes) {
		endIdx = len(allNodes)
	}
	paginatedNodes := allNodes[startIdx:endIdx]

	// Convert to pointers
	result := make([]*domain.Node, 0, len(paginatedNodes))
	for i := range paginatedNodes {
		result = append(result, &paginatedNodes[i])
	}

	return result, totalCount, nil
}

func (r *nodeRepository) Update(ctx context.Context, node *domain.Node) error {
	filter := bson.M{"_id": node.ID}
	update := bson.M{
		"$set": bson.M{
			"encrypted_readme":           node.EncryptedReadme,
			"encrypted_readme_signature": node.EncryptedReadmeSignature,
			"encrypted_dict":             node.EncryptedDict,
			"encrypted_dict_signature":   node.EncryptedDictSignature,
		},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *nodeRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}

func (r *nodeRepository) DeleteByDiagramID(ctx context.Context, diagramID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"diagram_id": diagramID})
	return err
}
