package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
	"relay/internal/token"
	"relay/internal/validation"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users        repositories.UserRepository
	refreshToken repositories.RefreshTokenRepository
	tokenService *token.Service
}

type LoginResult struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

const refreshTokenTTL = 30 * 24 * time.Hour

func NewUserService(
	users repositories.UserRepository,
	tokenService *token.Service,
	refreshToken *repositories.RefreshTokenRepository,
) *UserService {
	return &UserService{
		users:        users,
		refreshToken: *refreshToken,
		tokenService: tokenService,
	}
}

func (s *UserService) Register(ctx context.Context, name string, email string, password string) (*models.User, error) {
	// validation
	if err := validation.ValidateRegistraion(name, email, password); err != nil {
		return nil, err
	}

	// normalize email
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// create user
	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	// save user
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email string, password string) (*LoginResult, error) {
	// validation
	if err := validation.ValidateLogin(email, password); err != nil {
		return nil, err
	}

	// normalize email
	email = strings.TrimSpace(strings.ToLower(email))

	// find the user by calling repository
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, err
	}

	//compared hashed stored password with password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	//Generate access token from the TokenService
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	//Generate refresh token from the TokenService
	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	//Hash the refresh token using SHA-256
	sum := sha256.Sum256([]byte(refreshToken))

	//Convert to string
	hash := hex.EncodeToString(sum[:])

	//Build the model
	refresh := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}

	//Save the model
	if err := s.refreshToken.Create(ctx, refresh); err != nil {
		return nil, err
	}

	//return user in a new loginresult struct
	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Profile(ctx context.Context, userID int64) (*models.User, error) {
	return s.users.GetByID(ctx, userID)
}
