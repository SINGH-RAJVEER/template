package main

import (
	"context"
	"time"

	"template/database"
)

type user = database.User
type session = database.Session
type authSession = database.AuthSession
type sessionMetadata = database.SessionMetadata

var (
	errNotFound  = database.ErrNotFound
	errEmailUsed = database.ErrEmailUsed
)

type authStore interface {
	Register(context.Context, string, string, string, sessionMetadata, time.Duration) (authSession, string, error)
	Credentials(context.Context, string) (user, string, error)
	CreateSession(context.Context, user, sessionMetadata, time.Duration) (authSession, string, error)
	Session(context.Context, string) (authSession, error)
	DeleteSession(context.Context, string) error
}
