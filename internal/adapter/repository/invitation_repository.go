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

type invitationRepository struct {
	model mgod.EntityMongoModel[domain.Invitation]
}

func NewInvitationRepository(collectionName string) (port.InvitationRepository, error) {
	opts := schemaopt.SchemaOptions{
		Collection: collectionName,
		Timestamps: true,
	}
	model, err := mgod.NewEntityMongoModel(domain.Invitation{}, opts)
	if err != nil {
		return nil, err
	}

	return &invitationRepository{model: model}, nil
}

func (r *invitationRepository) Create(ctx context.Context, invitation *domain.Invitation) (*domain.Invitation, error) {
	result, err := r.model.InsertOne(ctx, *invitation)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *invitationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Invitation, error) {
	return r.model.FindOne(ctx, bson.M{"_id": id})
}

func (r *invitationRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID, offset, limit int) ([]*domain.Invitation, int64, error) {
	filter := bson.M{"project_id": projectID}

	allInvitations, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allInvitations))

	// Apply pagination manually
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allInvitations) {
		return []*domain.Invitation{}, totalCount, nil
	}
	if endIdx > len(allInvitations) {
		endIdx = len(allInvitations)
	}
	paginated := allInvitations[startIdx:endIdx]

	result := make([]*domain.Invitation, 0, len(paginated))
	for i := range paginated {
		result = append(result, &paginated[i])
	}

	return result, totalCount, nil
}

func (r *invitationRepository) FindByInviteeID(ctx context.Context, inviteeUserID primitive.ObjectID, offset, limit int) ([]*domain.Invitation, int64, error) {
	filter := bson.M{"invitee_user_id": inviteeUserID, "status": domain.InvitationStatusPending}

	allInvitations, err := r.model.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	totalCount := int64(len(allInvitations))

	// Apply pagination manually
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= len(allInvitations) {
		return []*domain.Invitation{}, totalCount, nil
	}
	if endIdx > len(allInvitations) {
		endIdx = len(allInvitations)
	}
	paginated := allInvitations[startIdx:endIdx]

	result := make([]*domain.Invitation, 0, len(paginated))
	for i := range paginated {
		result = append(result, &paginated[i])
	}

	return result, totalCount, nil
}

func (r *invitationRepository) FindByProjectAndInvitee(ctx context.Context, projectID, inviteeUserID primitive.ObjectID) (*domain.Invitation, error) {
	filter := bson.M{
		"project_id":      projectID,
		"invitee_user_id": inviteeUserID,
		"status":          domain.InvitationStatusPending,
	}
	return r.model.FindOne(ctx, filter)
}

func (r *invitationRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	_, err := r.model.UpdateMany(ctx, bson.M{"_id": invitation.ID}, bson.M{
		"$set": bson.M{
			"status": invitation.Status,
		},
	})
	return err
}

func (r *invitationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.model.DeleteMany(ctx, bson.M{"_id": id})
	return err
}
