package service

import (
	"context"
	"errors"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrCurrentPasswordWrong  = errors.New("current password is incorrect")
	ErrSamePassword          = errors.New("new password must be different")
)

type UserService struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	argon2Params     *Argon2Params
}

func NewUserService(
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	argon2Params *Argon2Params,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		argon2Params:     argon2Params,
	}
}

// GetUserProfile retrieves user profile by ID
func (s *UserService) GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, req dto.UpdateProfileRequest) (*domain.User, error) {
	// Get current user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil && *req.Email != user.Email {
		// Check if email already exists
		exists, err := s.userRepo.ExistsByEmail(ctx, *req.Email, userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = *req.Email
	}

	if req.Username != nil && *req.Username != user.Username {
		// Check if username already exists
		exists, err := s.userRepo.ExistsByUsername(ctx, *req.Username, userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUsernameAlreadyExists
		}
		user.Username = *req.Username
	}

	// Update user in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID primitive.ObjectID, currentPassword, newPassword string) error {
	// Get current user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	match, err := ComparePassword(currentPassword, user.Password)
	if err != nil {
		return err
	}
	if !match {
		return ErrCurrentPasswordWrong
	}

	// Check if new password is different
	sameAsOld, err := ComparePassword(newPassword, user.Password)
	if err != nil {
		return err
	}
	if sameAsOld {
		return ErrSamePassword
	}

	// Hash new password
	hashedPassword, err := HashPassword(newPassword, s.argon2Params)
	if err != nil {
		return err
	}

	// Update password
	user.Password = hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all refresh tokens for security
	if err := s.refreshTokenRepo.RevokeByUserID(ctx, userID); err != nil {
		// Log error but don't fail the password change
		// In a production app, you'd want proper logging here
		return nil
	}

	return nil
}
