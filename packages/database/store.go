package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrEmailUsed = errors.New("email is already registered")
)

//go:embed schema.sql
var schemaFiles embed.FS

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	schema, err := schemaFiles.ReadFile("schema.sql")
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("read embedded schema: %w", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		pool.Close()
		return nil, fmt.Errorf("initialize database schema: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

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

func (s *Store) Credentials(ctx context.Context, email string) (User, string, error) {
	var result User
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.email_verified, u.image, u.created_at, u.updated_at, a.password
		FROM "user" u
		JOIN account a ON a.user_id = u.id AND a.provider_id = 'credential'
		WHERE u.email = $1
	`, email).Scan(
		&result.ID,
		&result.Name,
		&result.Email,
		&result.EmailVerified,
		&result.Image,
		&result.CreatedAt,
		&result.UpdatedAt,
		&passwordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	return result, passwordHash, err
}

func (s *Store) CreateSession(
	ctx context.Context,
	account User,
	metadata SessionMetadata,
	ttl time.Duration,
) (AuthSession, string, error) {
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

func insertSession(
	ctx context.Context,
	db rowExecutor,
	account User,
	metadata SessionMetadata,
	ttl time.Duration,
) (AuthSession, string, error) {
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
	`,
		newSession.ID,
		newSession.ExpiresAt,
		hashToken(token),
		now,
		newSession.IPAddress,
		newSession.UserAgent,
		newSession.UserID,
	)
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

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
