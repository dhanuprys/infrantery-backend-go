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

type projectRepository struct {
	model mgod.EntityMongoModel[domain.Project]
}

func NewProjectRepository(collectionName string) (port.ProjectRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.Project{}, opts)
	if err != nil {
		return nil, err
	}

	return &projectRepository{model: model}, nil
}

func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	result, err := r.model.InsertOne(ctx, *project)
	if err != nil {
		return err
	}
	project.ID = result.ID
	return nil
}

func (r *projectRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Project, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *projectRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, offset, limit int) ([]*domain.Project, int64, error) {
	// First, get all project IDs that the user is a member of
	memberOpts := schemaopt.SchemaOptions{
		Collection: "project_members",
		Timestamps: false,
	}
	memberModel, err := mgod.NewEntityMongoModel(domain.ProjectMember{}, memberOpts)
	if err != nil {
		return nil, 0, err
	}

	members, err := memberModel.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, 0, err
	}

	if len(members) == 0 {
		return []*domain.Project{}, 0, nil
	}

	// Extract project IDs
	projectIDs := make([]primitive.ObjectID, 0, len(members))
	for _, member := range members {
		projectIDs = append(projectIDs, member.ProjectID)
	}

	// Get total count
	totalCount := int64(len(projectIDs))

	// Apply pagination to project IDs
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(projectIDs) {
		return []*domain.Project{}, totalCount, nil
	}
	if endIdx > len(projectIDs) {
		endIdx = len(projectIDs)
	}
	paginatedIDs := projectIDs[startIdx:endIdx]

	// Find projects with paginated IDs
	projects, err := r.model.Find(ctx, bson.M{"_id": bson.M{"$in": paginatedIDs}})
	if err != nil {
		return nil, 0, err
	}

	// Convert to pointers
	result := make([]*domain.Project, 0, len(projects))
	for i := range projects {
		result = append(result, &projects[i])
	}

	return result, totalCount, nil
}

func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	filter := bson.M{"_id": project.ID}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "name", Value: project.Name},
			{Key: "description", Value: project.Description},
			{Key: "key_epoch", Value: project.KeyEpoch},
		}},
	}
	_, err := r.model.UpdateMany(ctx, filter, update)
	return err
}

func (r *projectRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}
