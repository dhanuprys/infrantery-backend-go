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

type nodeVaultRepository struct {
	model mgod.EntityMongoModel[domain.NodeVault]
}

func NewNodeVaultRepository(collectionName string) (port.NodeVaultRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.NodeVault{}, opts)
	if err != nil {
		return nil, err
	}

	return &nodeVaultRepository{model: model}, nil
}

func (r *nodeVaultRepository) Create(ctx context.Context, vault *domain.NodeVault) error {
	result, err := r.model.InsertOne(ctx, *vault)
	if err != nil {
		return err
	}
	vault.ID = result.ID
	return nil
}

func (r *nodeVaultRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.NodeVault, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *nodeVaultRepository) FindByNodeID(ctx context.Context, nodeID primitive.ObjectID) ([]*domain.NodeVault, error) {
	// Find returns []T, we need []*T
	vaults, err := r.model.Find(ctx, bson.M{"node_id": nodeID})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NodeVault, 0, len(vaults))
	for i := range vaults {
		result = append(result, &vaults[i])
	}
	return result, nil
}

func (r *nodeVaultRepository) Update(ctx context.Context, vault *domain.NodeVault) error {
	filter := bson.M{"_id": vault.ID}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "encrypted_value", Value: vault.EncryptedValue},
			{Key: "encrypted_value_signature", Value: vault.EncryptedValueSignature},
		}},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *nodeVaultRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}

func (r *nodeVaultRepository) DeleteByNodeID(ctx context.Context, nodeID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"node_id": nodeID})
	return err
}
