package port

import (
	"context"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	ExistsByEmail(ctx context.Context, email string, excludeUserID primitive.ObjectID) (bool, error)
	ExistsByUsername(ctx context.Context, username string, excludeUserID primitive.ObjectID) (bool, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*domain.User, error)
}

type InvitationRepository interface {
	Create(ctx context.Context, invitation *domain.Invitation) (*domain.Invitation, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Invitation, error)
	FindByProjectID(ctx context.Context, projectID primitive.ObjectID, offset, limit int) ([]*domain.Invitation, int64, error)
	FindByInviteeID(ctx context.Context, inviteeUserID primitive.ObjectID, offset, limit int) ([]*domain.Invitation, int64, error)
	FindByProjectAndInvitee(ctx context.Context, projectID, inviteeUserID primitive.ObjectID) (*domain.Invitation, error)
	Update(ctx context.Context, invitation *domain.Invitation) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Project, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID, offset, limit int) ([]*domain.Project, int64, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type ProjectMemberRepository interface {
	Create(ctx context.Context, member *domain.ProjectMember) error
	FindByProjectID(ctx context.Context, projectID primitive.ObjectID, offset, limit int) ([]*domain.ProjectMember, int64, error)
	FindByProjectAndUser(ctx context.Context, projectID, userID primitive.ObjectID) (*domain.ProjectMember, error)
	Update(ctx context.Context, member *domain.ProjectMember) error
	Delete(ctx context.Context, projectID, userID primitive.ObjectID) error
	DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error
}

type NoteRepository interface {
	Create(ctx context.Context, note *domain.Note) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Note, error)
	FindByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Note, error)
	Update(ctx context.Context, note *domain.Note) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error
}

type DiagramRepository interface {
	Create(ctx context.Context, diagram *domain.Diagram) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Diagram, error)
	FindByProjectID(ctx context.Context, projectID primitive.ObjectID, rootOnly bool, offset, limit int) ([]*domain.Diagram, int64, error)
	FindAllByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Diagram, error)
	Update(ctx context.Context, diagram *domain.Diagram) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByProjectID(ctx context.Context, projectID primitive.ObjectID) error
}

type NodeRepository interface {
	Create(ctx context.Context, node *domain.Node) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Node, error)
	FindByDiagramID(ctx context.Context, diagramID primitive.ObjectID, offset, limit int) ([]*domain.Node, int64, error)
	FindByDiagramIDs(ctx context.Context, diagramIDs []primitive.ObjectID) ([]*domain.Node, error)
	Update(ctx context.Context, node *domain.Node) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByDiagramID(ctx context.Context, diagramID primitive.ObjectID) error
}

type NodeVaultRepository interface {
	Create(ctx context.Context, vault *domain.NodeVault) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.NodeVault, error)
	FindByNodeID(ctx context.Context, nodeID primitive.ObjectID) ([]*domain.NodeVault, error)
	FindByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.NodeVault, error)
	Update(ctx context.Context, vault *domain.NodeVault) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByNodeID(ctx context.Context, nodeID primitive.ObjectID) error
}
