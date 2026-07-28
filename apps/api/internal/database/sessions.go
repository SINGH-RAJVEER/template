package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateSession(ctx context.Context, account User, metadata SessionMetadata, ttl time.Duration) (AuthSession, string, error) {
	return insertSession(ctx, s.pool, account, metadata, ttl)
}

func (s *Store) Session(ctx context.Context, token string) (AuthSession, error) {
	var result AuthSession
	err := s.pool.QueryRow(ctx, `
		SELECT
			s.id, s.user_id, s.expires_at, s.ip_address, s.user_agent, s.created_at, s.updated_at,
			u.id, u.name, u.email, u.email_verified, u.image, u.created_at, u.updated_at
		FROM session s
		JOIN "user" u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, hashToken(token)).Scan(
		&result.Session.ID,
		&result.Session.UserID,
		&result.Session.ExpiresAt,
		&result.Session.IPAddress,
		&result.Session.UserAgent,
		&result.Session.CreatedAt,
		&result.Session.UpdatedAt,
		&result.User.ID,
		&result.User.Name,
		&result.User.Email,
		&result.User.EmailVerified,
		&result.User.Image,
		&result.User.CreatedAt,
		&result.User.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthSession{}, ErrNotFound
	}
	return result, err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM session WHERE token = $1`, hashToken(token))
	return err
}

type rowExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertSession(ctx context.Context, db rowExecutor, account User, metadata SessionMetadata, ttl time.Duration) (AuthSession, string, error) {
	token, err := randomToken(32)
	if err != nil {
		return AuthSession{}, "", err
	}

	now := time.Now().UTC()
	newSession := Session{
		ID:        randomID(),
		UserID:    account.ID,
		ExpiresAt: now.Add(ttl),
		IPAddress: optionalString(metadata.IPAddress),
		UserAgent: optionalString(metadata.UserAgent),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = db.Exec(ctx, `
		INSERT INTO session
			(id, expires_at, token, created_at, updated_at, ip_address, user_agent, user_id)
		VALUES ($1, $2, $3, $4, $4, $5, $6, $7)
	`, newSession.ID, newSession.ExpiresAt, hashToken(token), now, newSession.IPAddress, newSession.UserAgent, newSession.UserID)
	if err != nil {
		return AuthSession{}, "", err
	}
	return AuthSession{User: account, Session: newSession}, token, nil
}

func randomID() string {
	id, err := randomToken(18)
	if err != nil {
		panic(fmt.Sprintf("generate random ID: %v", err))
	}
	return id
}

func randomToken(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
