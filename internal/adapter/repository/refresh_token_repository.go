package repository

import (
	"context"
	"time"

	"github.com/Lyearn/mgod"
	"github.com/Lyearn/mgod/schema/schemaopt"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type refreshTokenRepository struct {
	model mgod.EntityMongoModel[domain.RefreshToken]
}

func NewRefreshTokenRepository(collection string) (port.RefreshTokenRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collection,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.RefreshToken{}, opts)
	if err != nil {
		return nil, err
	}

	return &refreshTokenRepository{model: model}, nil
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.model.InsertOne(ctx, *token)
	return err
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	result, err := r.model.FindOne(ctx, bson.M{"token": token, "is_revoked": false})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.model.UpdateMany(ctx, bson.M{"user_id": userID}, bson.M{"$set": bson.M{"is_revoked": true}})
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": time.Now()}})
	return err
}
