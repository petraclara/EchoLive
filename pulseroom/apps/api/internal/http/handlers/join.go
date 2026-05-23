package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/models"
	"github.com/pulseroom/api/internal/repository"
)

type JoinHandler struct {
	Store     *repository.Store
	APIURL    string
	WebAppURL string
}

func (h *JoinHandler) Join(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		Error(w, http.StatusBadRequest, "code required")
		return
	}
	e, err := h.Store.GetEventByJoinCode(r.Context(), req.Code)
	if err == repository.ErrNotFound {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "join failed")
		return
	}
	ua := r.UserAgent()
	_, token, err := h.Store.CreateAttendeeSession(r.Context(), e.ID, ua)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not join")
		return
	}
	bootstrap, err := h.buildPublic(r, e.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not load event")
		return
	}
	JSON(w, http.StatusOK, map[string]any{
		"event_id":       e.ID,
		"session_token":  token,
		"event":          bootstrap,
		"ws_url":         h.APIURL + "/ws/events/" + e.ID.String() + "?token=" + token,
		"attendee_path":  "/e/" + e.ID.String(),
	})
}

func (h *JoinHandler) Public(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	pub, err := h.buildPublic(r, eventID)
	if err == repository.ErrNotFound {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not load event")
		return
	}
	JSON(w, http.StatusOK, pub)
}

func (h *JoinHandler) buildPublic(r *http.Request, eventID uuid.UUID) (*models.PublicEvent, error) {
	e, err := h.Store.GetEvent(r.Context(), eventID)
	if err != nil {
		return nil, err
	}
	ann, err := h.Store.ListAnnouncements(r.Context(), eventID, nil, 50)
	if err != nil {
		return nil, err
	}
	pinned, err := h.Store.GetPinnedAnnouncement(r.Context(), eventID)
	if err != nil {
		return nil, err
	}
	resources, err := h.Store.ListResources(r.Context(), eventID)
	if err != nil {
		return nil, err
	}
	agenda, err := h.Store.ListAgenda(r.Context(), eventID)
	if err != nil {
		return nil, err
	}
	count, _ := h.Store.AttendeeCount(r.Context(), eventID, 2*time.Minute)
	return &models.PublicEvent{
		Event:         *e,
		Pinned:        pinned,
		Announcements: ann,
		Resources:     resources,
		Agenda:        agenda,
		AttendeeCount: count,
	}, nil
}
