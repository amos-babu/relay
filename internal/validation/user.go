package validation

import (
	"errors"
	"strings"
)

func ValidateRegistraion(name, email, password string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if name == "" {
		return errors.New("name is required")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}
