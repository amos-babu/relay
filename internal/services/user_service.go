package services

import (
	"context"
	"relay/internal/repositories"
	"relay/internal/validation"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Register(ctx context.Context, name string, email string, password string) error {
	if err := validation.ValidateRegistraion(name, email, password); err != nil {
		return err
	}
	return nil
}
