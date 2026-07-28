package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "session_token"

type apiServer struct {
	store        authStore
	webURL       string
	cookieSecure bool
	sessionTTL   time.Duration
}

func newAPIServer(store authStore, cfg config) http.Handler {
	server := &apiServer{
		store:        store,
		webURL:       strings.TrimRight(cfg.webURL, "/"),
		cookieSecure: cfg.cookieSecure,
		sessionTTL:   cfg.sessionTTL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /api/auth/sign-up/email", server.signUp)
	mux.HandleFunc("POST /api/auth/sign-in/email", server.signIn)
	mux.HandleFunc("POST /api/auth/sign-out", server.signOut)
	mux.HandleFunc("GET /api/auth/session", server.getSession)
	mux.HandleFunc("GET /api/me", server.me)

	return server.recoverPanic(server.logRequests(server.cors(mux)))
}

func (s *apiServer) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) signUp(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		CallbackURL string `json:"callbackURL"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if message := validateCredentials(input.Name, input.Email, input.Password); message != "" {
		writeError(w, http.StatusBadRequest, "INVALID_CREDENTIALS", message)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create account")
		return
	}

	result, token, err := s.store.Register(
		r.Context(),
		input.Name,
		input.Email,
		string(passwordHash),
		requestMetadata(r),
		s.sessionTTL,
	)
	if errors.Is(err, errEmailUsed) {
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Email is already registered")
		return
	}
	if err != nil {
		slog.Error("register user", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create account")
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, result)
}

func (s *apiServer) signIn(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		CallbackURL string `json:"callbackURL"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	account, passwordHash, err := s.store.Credentials(r.Context(), input.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_EMAIL_OR_PASSWORD", "Invalid email or password")
		return
	}

	result, token, err := s.store.CreateSession(
		r.Context(),
		account,
		requestMetadata(r),
		s.sessionTTL,
	)
	if err != nil {
		slog.Error("create session", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not sign in")
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, result)
}

func (s *apiServer) signOut(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if err := s.store.DeleteSession(r.Context(), cookie.Value); err != nil {
			slog.Error("delete session", "error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not sign out")
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *apiServer) getSession(w http.ResponseWriter, r *http.Request) {
	result, ok := s.authenticatedSession(r)
	if !ok {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *apiServer) me(w http.ResponseWriter, r *http.Request) {
	result, ok := s.authenticatedSession(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]user{"user": result.User})
}

func (s *apiServer) authenticatedSession(r *http.Request) (authSession, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return authSession{}, false
	}

	result, err := s.store.Session(r.Context(), cookie.Value)
	if err != nil {
		if !errors.Is(err, errNotFound) {
			slog.Error("load session", "error", err)
		}
		return authSession{}, false
	}
	return result, true
}

func (s *apiServer) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *apiServer) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if origin != "" && origin != s.webURL {
			writeError(w, http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "Origin is not allowed")
			return
		}

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *apiServer) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		response := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(response, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", response.status,
			"duration", time.Since(started),
		)
	})
}

func (s *apiServer) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic serving request", "value", recovered, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func decodeJSON(r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); errors.Is(err, io.EOF) {
		return nil
	}
	return fmt.Errorf("request body must contain one JSON object")
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

func requestMetadata(r *http.Request) sessionMetadata {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ipAddress = r.RemoteAddr
	}
	return sessionMetadata{IPAddress: ipAddress, UserAgent: r.UserAgent()}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "message": message})
}
