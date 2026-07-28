package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"template/api/internal/database"
	apihttp "template/api/internal/http"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeStore struct {
	account      database.User
	passwordHash string
	session      database.AuthSession
	token        string
	registerErr  error
	deletedToken string
}

func (s *fakeStore) Register(_ context.Context, name, email, passwordHash string, _ database.SessionMetadata, _ time.Duration) (database.AuthSession, string, error) {
	if s.registerErr != nil {
		return database.AuthSession{}, "", s.registerErr
	}
	s.passwordHash = passwordHash
	s.session.User.Name = name
	s.session.User.Email = email
	return s.session, s.token, nil
}

func (s *fakeStore) Credentials(_ context.Context, email string) (database.User, string, error) {
	if email != s.account.Email {
		return database.User{}, "", database.ErrNotFound
	}
	return s.account, s.passwordHash, nil
}

func (s *fakeStore) CreateSession(_ context.Context, _ database.User, _ database.SessionMetadata, _ time.Duration) (database.AuthSession, string, error) {
	return s.session, s.token, nil
}

func (s *fakeStore) Session(_ context.Context, token string) (database.AuthSession, error) {
	if token != s.token {
		return database.AuthSession{}, database.ErrNotFound
	}
	return s.session, nil
}

func (s *fakeStore) DeleteSession(_ context.Context, token string) error {
	s.deletedToken = token
	return nil
}

func testServer(store Store) http.Handler {
	handler := NewHandler(store, Config{SessionTTL: 24 * time.Hour})
	return apihttp.New(handler, "http://localhost:3000")
}

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	testServer(&fakeStore{}).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestSignUpSetsSessionCookie(t *testing.T) {
	store := &fakeStore{
		session: database.AuthSession{User: database.User{ID: "user-1"}, Session: database.Session{ID: "session-1"}},
		token:   "raw-session-token",
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(`{"name":"Ada","email":"ADA@example.com","password":"password123"}`))
	response := httptest.NewRecorder()
	testServer(store).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if store.session.User.Email != "ada@example.com" {
		t.Fatalf("expected normalized email, got %q", store.session.User.Email)
	}
	if bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte("password123")) != nil {
		t.Fatal("password was not stored as a bcrypt hash")
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != SessionCookieName {
		t.Fatalf("expected session cookie, got %#v", cookies)
	}
	if !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("session cookie is missing security attributes")
	}
}

func TestSignUpRejectsTrailingJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"password123"} {}`))
	response := httptest.NewRecorder()
	testServer(&fakeStore{}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestSignInRejectsInvalidPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{account: database.User{ID: "user-1", Email: "ada@example.com"}, passwordHash: string(passwordHash)}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in/email", bytes.NewBufferString(`{"email":"ada@example.com","password":"wrong-password"}`))
	response := httptest.NewRecorder()
	testServer(store).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestMeRequiresValidSession(t *testing.T) {
	store := &fakeStore{session: database.AuthSession{User: database.User{ID: "user-1", Name: "Ada"}}, token: "valid-token"}

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	unauthorizedResponse := httptest.NewRecorder()
	testServer(store).ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing cookie to return 401, got %d", unauthorizedResponse.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	authorizedRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: store.token})
	authorizedResponse := httptest.NewRecorder()
	testServer(store).ServeHTTP(authorizedResponse, authorizedRequest)
	if authorizedResponse.Code != http.StatusOK {
		t.Fatalf("expected valid session to return 200, got %d", authorizedResponse.Code)
	}
	var payload struct {
		User database.User `json:"user"`
	}
	if err := json.NewDecoder(authorizedResponse.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.User.ID != store.session.User.ID {
		t.Fatalf("expected user %q, got %q", store.session.User.ID, payload.User.ID)
	}
}

func TestSignOutDeletesSession(t *testing.T) {
	store := &fakeStore{token: "valid-token"}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-out", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: store.token})
	response := httptest.NewRecorder()
	testServer(store).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if store.deletedToken != store.token {
		t.Fatalf("expected token %q to be deleted, got %q", store.token, store.deletedToken)
	}
	if cookies := response.Result().Cookies(); len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatal("expected the session cookie to be expired")
	}
}

func TestCORS(t *testing.T) {
	server := testServer(&fakeStore{})
	allowedRequest := httptest.NewRequest(http.MethodOptions, "/api/auth/session", nil)
	allowedRequest.Header.Set("Origin", "http://localhost:3000")
	allowedResponse := httptest.NewRecorder()
	server.ServeHTTP(allowedResponse, allowedRequest)
	if allowedResponse.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status 204, got %d", allowedResponse.Code)
	}
	if allowedResponse.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentialed CORS response")
	}

	blockedRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	blockedRequest.Header.Set("Origin", "https://attacker.example")
	blockedResponse := httptest.NewRecorder()
	server.ServeHTTP(blockedResponse, blockedRequest)
	if blockedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected untrusted origin status 403, got %d", blockedResponse.Code)
	}
}

func TestEmailConflict(t *testing.T) {
	store := &fakeStore{registerErr: database.ErrEmailUsed}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"password123"}`))
	response := httptest.NewRecorder()
	testServer(store).ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", response.Code)
	}
}

var _ Store = (*fakeStore)(nil)
