package auth

import (
	"context"
	"template/api/internal/database"
	"time"
)

type Store interface {
	Register(context.Context, string, string, string, database.SessionMetadata, time.Duration) (database.AuthSession, string, error)
	Credentials(context.Context, string) (database.User, string, error)
	CreateSession(context.Context, database.User, database.SessionMetadata, time.Duration) (database.AuthSession, string, error)
	Session(context.Context, string) (database.AuthSession, error)
	DeleteSession(context.Context, string) error
}
