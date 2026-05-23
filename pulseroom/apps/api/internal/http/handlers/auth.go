package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pulseroom/api/internal/auth"
	"github.com/pulseroom/api/internal/http/middleware"
	"github.com/pulseroom/api/internal/repository"
)

type AuthHandler struct {
	Store     *repository.Store
	JWTSecret string
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 8 || req.Name == "" {
		Error(w, http.StatusBadRequest, "email, name, and password (8+ chars) required")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	o, err := h.Store.CreateOrganizer(r.Context(), req.Email, hash, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			Error(w, http.StatusConflict, "email already registered")
			return
		}
		Error(w, http.StatusInternalServerError, "could not create account")
		return
	}
	token, err := auth.IssueToken(h.JWTSecret, o.ID, o.Email)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	JSON(w, http.StatusCreated, map[string]any{"organizer": o, "token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	id, hash, name, err := h.Store.GetOrganizerByEmail(r.Context(), email)
	if err == repository.ErrNotFound || !auth.CheckPassword(hash, req.Password) {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "login failed")
		return
	}
	token, err := auth.IssueToken(h.JWTSecret, id, email)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	JSON(w, http.StatusOK, map[string]any{
		"organizer": map[string]any{"id": id, "email": email, "name": name},
		"token":     token,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	id := middleware.OrganizerID(r)
	o, err := h.Store.GetOrganizer(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "organizer not found")
		return
	}
	JSON(w, http.StatusOK, o)
}
