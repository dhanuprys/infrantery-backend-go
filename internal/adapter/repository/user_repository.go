package repository

import (
	"context"

	"github.com/Lyearn/mgod"
	"github.com/Lyearn/mgod/schema/schemaopt"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	model mgod.EntityMongoModel[domain.User]
}

func NewUserRepository(collection string) (port.UserRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collection,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.User{}, opts)
	if err != nil {
		return nil, err
	}

	return &userRepository{model: model}, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := r.model.InsertOne(ctx, *user)
	if err != nil {
		return err
	}
	// Update user with generated ID
	user.ID = result.ID
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	result, err := r.model.FindOne(ctx, bson.M{"email": email})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	result, err := r.model.FindOne(ctx, bson.M{"username": username})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	result, err := r.model.FindOne(ctx, bson.M{"_id": id})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.model.UpdateMany(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"name":     user.Name,
			"username": user.Username,
			"email":    user.Email,
			"password": user.Password,
		},
	})
	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string, excludeUserID primitive.ObjectID) (bool, error) {
	result, err := r.model.FindOne(ctx, bson.M{
		"email": email,
		"_id":   bson.M{"$ne": excludeUserID},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return result != nil, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string, excludeUserID primitive.ObjectID) (bool, error) {
	result, err := r.model.FindOne(ctx, bson.M{
		"username": username,
		"_id":      bson.M{"$ne": excludeUserID},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return result != nil, nil
}
