package auth

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"template/api/internal/database"
	apihttp "template/api/internal/http"
	"time"
)

const SessionCookieName = "session_token"

type Config struct {
	CookieSecure bool
	SessionTTL   time.Duration
}

type Handler struct {
	service      service
	cookieSecure bool
}

func NewHandler(store Store, cfg Config) *Handler {
	return &Handler{
		service:      service{store: store, sessionTTL: cfg.SessionTTL},
		cookieSecure: cfg.CookieSecure,
	}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var input signUpRequest
	if err := apihttp.DecodeJSON(r, &input); err != nil {
		apihttp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	result, token, err := h.service.register(r.Context(), input.Name, input.Email, input.Password, requestMetadata(r))
	var invalid *validationError
	switch {
	case errors.As(err, &invalid):
		apihttp.WriteError(w, http.StatusBadRequest, "INVALID_CREDENTIALS", invalid.Error())
		return
	case errors.Is(err, database.ErrEmailUsed):
		apihttp.WriteError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Email is already registered")
		return
	case err != nil:
		slog.Error("register user", "error", err)
		apihttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create account")
		return
	}

	h.setSessionCookie(w, token)
	apihttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var input signInRequest
	if err := apihttp.DecodeJSON(r, &input); err != nil {
		apihttp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	result, token, err := h.service.signIn(r.Context(), input.Email, input.Password, requestMetadata(r))
	if errors.Is(err, ErrInvalidCredentials) {
		apihttp.WriteError(w, http.StatusUnauthorized, "INVALID_EMAIL_OR_PASSWORD", "Invalid email or password")
		return
	}
	if err != nil {
		slog.Error("create session", "error", err)
		apihttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not sign in")
		return
	}

	h.setSessionCookie(w, token)
	apihttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if err := h.service.store.DeleteSession(r.Context(), cookie.Value); err != nil {
			slog.Error("delete session", "error", err)
			apihttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not sign out")
			return
		}
	}

	http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode})
	apihttp.WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	result, ok := h.authenticatedSession(r)
	if !ok {
		apihttp.WriteJSON(w, http.StatusOK, nil)
		return
	}
	apihttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	result, ok := h.authenticatedSession(r)
	if !ok {
		apihttp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	apihttp.WriteJSON(w, http.StatusOK, map[string]database.User{"user": result.User})
}

func (h *Handler) authenticatedSession(r *http.Request) (database.AuthSession, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return database.AuthSession{}, false
	}
	result, err := h.service.store.Session(r.Context(), cookie.Value)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			slog.Error("load session", "error", err)
		}
		return database.AuthSession{}, false
	}
	return result, true
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: token, Path: "/", MaxAge: int(h.service.sessionTTL.Seconds()), HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode})
}

func requestMetadata(r *http.Request) database.SessionMetadata {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ipAddress = r.RemoteAddr
	}
	return database.SessionMetadata{IPAddress: ipAddress, UserAgent: r.UserAgent()}
}
