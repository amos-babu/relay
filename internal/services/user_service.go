package services

import (
	"context"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
	"relay/internal/token"
	"relay/internal/validation"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo         repositories.UserRepository
	refreshToken repositories.RefreshTokenRepository
	tokenService *token.Service
}

type LoginResult struct {
	User  *models.User
	Token string
}

func NewUserService(repo repositories.UserRepository, tokenService *token.Service) *UserService {
	return &UserService{
		repo:         repo,
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
	if err := s.repo.Create(ctx, user); err != nil {
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
	user, err := s.repo.GetByEmail(ctx, email)
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

	token, err := s.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	//return user in a new loginresult struct
	return &LoginResult{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserService) Profile(ctx context.Context, userID int64) (*models.User, error) {
	return s.repo.GetByID(ctx, userID)
}
