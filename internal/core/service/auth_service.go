package service

import (
	"context"
	"errors"
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
)

var (
	ErrUserExists         = errors.New("user with this email or username already exists")
	ErrInvalidCredentials = errors.New("invalid email/username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	jwtService       *JWTService
	argon2Params     *Argon2Params
}

func NewAuthService(
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	jwtService *JWTService,
	argon2Params *Argon2Params,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		argon2Params:     argon2Params,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existingEmail, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, ErrUserExists
	}

	existingUsername, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUsername != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password, s.argon2Params)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens (user will have _id after insert)
	// For mgod, we need to get the inserted user or handle ID differently
	// Let's find the user by email to get the ID
	createdUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, createdUser)
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email or username
	var user *domain.User
	var err error

	// Try email first
	user, err = s.userRepo.FindByEmail(ctx, req.EmailOrUsername)
	if err != nil {
		return nil, err
	}

	// If not found, try username
	if user == nil {
		user, err = s.userRepo.FindByUsername(ctx, req.EmailOrUsername)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	match, err := ComparePassword(req.Password, user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user)
}

// RefreshAccessToken generates a new access token from a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (*dto.AuthResponse, error) {
	// Find refresh token
	refreshToken, err := s.refreshTokenRepo.FindByToken(ctx, refreshTokenString)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, ErrInvalidToken
	}

	// Check if expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	// Generate new access token (keep same refresh token)
	accessToken, err := s.jwtService.GenerateAccessToken(refreshToken.UserID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    s.jwtService.GetAccessExpirySeconds(),
	}, nil
}

// generateTokens creates access and refresh tokens for a user
func (s *AuthService) generateTokens(ctx context.Context, user *domain.User) (*dto.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshTokenString, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshExpiry()),
		IsRevoked: false,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    s.jwtService.GetAccessExpirySeconds(),
	}, nil
}
