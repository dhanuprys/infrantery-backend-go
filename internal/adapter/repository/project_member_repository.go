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

type projectMemberRepository struct {
	model mgod.EntityMongoModel[domain.ProjectMember]
}

func NewProjectMemberRepository(collectionName string) (port.ProjectMemberRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.ProjectMember{}, opts)
	if err != nil {
		return nil, err
	}

	return &projectMemberRepository{model: model}, nil
}

func (r *projectMemberRepository) Create(ctx context.Context, member *domain.ProjectMember) error {
	_, err := r.model.InsertOne(ctx, *member)
	return err
}

func (r *projectMemberRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID, offset, limit int) ([]*domain.ProjectMember, int64, error) {
	filter := bson.M{"project_id": projectID}

	// Get total count
	// Note: mgod doesn't expose CountDocuments, so we'll fetch all and count
	// For production, you might want to use the underlying collection directly
	allMembers, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allMembers))

	// Apply pagination manually (since mgod.Find doesn't support skip/limit)
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allMembers) {
		return []*domain.ProjectMember{}, totalCount, nil
	}
	if endIdx > len(allMembers) {
		endIdx = len(allMembers)
	}
	paginatedMembers := allMembers[startIdx:endIdx]

	// Convert to pointers
	result := make([]*domain.ProjectMember, 0, len(paginatedMembers))
	for i := range paginatedMembers {
		result = append(result, &paginatedMembers[i])
	}

	return result, totalCount, nil
}

func (r *projectMemberRepository) FindByProjectAndUser(ctx context.Context, projectID, userID primitive.ObjectID) (*domain.ProjectMember, error) {
	return r.model.FindOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	})
}

func (r *projectMemberRepository) Update(ctx context.Context, member *domain.ProjectMember) error {
	filter := bson.M{
		"project_id": member.ProjectID,
		"user_id":    member.UserID,
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "permissions", Value: member.Permissions},
			{Key: "role", Value: member.Role},
			{Key: "keyrings", Value: member.Keyrings},
		}},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *projectMemberRepository) Delete(ctx context.Context, projectID, userID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	})
	return err
}

func (r *projectMemberRepository) DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}
