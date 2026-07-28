package database

import (
	"context"
	"time"
)

func (s *Store) Register(
	ctx context.Context,
	name string,
	email string,
	passwordHash string,
	metadata SessionMetadata,
	ttl time.Duration,
) (AuthSession, string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AuthSession{}, "", err
	}
	defer tx.Rollback(ctx) // No-op after Commit.

	now := time.Now().UTC()
	newUser := User{
		ID:        randomID(),
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, $4, $4)
	`, newUser.ID, newUser.Name, newUser.Email, now)
	if isUniqueViolation(err) {
		return AuthSession{}, "", ErrEmailUsed
	}
	if err != nil {
		return AuthSession{}, "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO account (id, account_id, provider_id, user_id, password, created_at, updated_at)
		VALUES ($1, $2, 'credential', $3, $4, $5, $5)
	`, randomID(), newUser.Email, newUser.ID, passwordHash, now)
	if err != nil {
		return AuthSession{}, "", err
	}

	newSession, token, err := insertSession(ctx, tx, newUser, metadata, ttl)
	if err != nil {
		return AuthSession{}, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return AuthSession{}, "", err
	}
	return newSession, token, nil
}
