package validation

import (
	"fmt"
	"net/mail"
	"strings"
)

type ValidationError struct {
	Errors map[string]string `json:"errors"`
}

func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// ValidationError returns as error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d error(s)", len(v.Errors))
}

func ValidateRegistration(name, email, password string) error {
	vErr := &ValidationError{Errors: make(map[string]string)}

	//Name
	name = strings.TrimSpace(name)
	if name == "" {
		vErr.Errors["name"] = "name is required"
	} else if len(name) < 2 {
		vErr.Errors["name"] = "name must be at least 2 characters"
	}

	//Email
	email = strings.TrimSpace(email)
	if email == "" {
		vErr.Errors["email"] = "email is required"
	} else if _, err := mail.ParseAddress(email); err != nil {
		vErr.Errors["email"] = "invalid email address format"
	}

	//Password
	if password == "" {
		vErr.Errors["password"] = "password is required"
	} else if len(password) < 8 {
		vErr.Errors["password"] = "password must be at least 8 characters"
	}

	if vErr.HasErrors() {
		return vErr
	}

	return nil
}

func ValidateLogin(email, password string) error {
	vErr := &ValidationError{Errors: make(map[string]string)}

	//Email
	email = strings.TrimSpace(email)
	if email == "" {
		vErr.Errors["email"] = "email is required"
	} else if _, err := mail.ParseAddress(email); err != nil {
		vErr.Errors["email"] = "invalid email address format"
	}

	//Password
	if password == "" {
		vErr.Errors["password"] = "password is required"
	}

	if vErr.HasErrors() {
		return vErr
	}

	return nil
}
