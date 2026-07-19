package services

import (
	"context"
	"relay/internal/models"
	"relay/internal/repositories"
	"relay/internal/validation"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{
		repo: repo,
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
