package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"template/api/internal/database"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type service struct {
	store      Store
	sessionTTL time.Duration
}

func (s *service) register(ctx context.Context, name, email, password string, metadata database.SessionMetadata) (database.AuthSession, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if message := validateCredentials(name, email, password); message != "" {
		return database.AuthSession{}, "", &validationError{message: message}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.AuthSession{}, "", err
	}
	return s.store.Register(ctx, name, email, string(passwordHash), metadata, s.sessionTTL)
}

func (s *service) signIn(ctx context.Context, email, password string, metadata database.SessionMetadata) (database.AuthSession, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	account, passwordHash, err := s.store.Credentials(ctx, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return database.AuthSession{}, "", ErrInvalidCredentials
	}
	return s.store.CreateSession(ctx, account, metadata, s.sessionTTL)
}

func validateCredentials(name, email, password string) string {
	if name == "" || len(name) > 100 {
		return "Name must be between 1 and 100 characters"
	}
	address, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(address.Address, email) || len(email) > 254 {
		return "Enter a valid email address"
	}
	if len(password) < 8 || len(password) > 72 {
		return "Password must be between 8 and 72 characters"
	}
	return ""
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
