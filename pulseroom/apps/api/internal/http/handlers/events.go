package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/http/middleware"
	"github.com/pulseroom/api/internal/models"
	"github.com/pulseroom/api/internal/repository"
)

type EventHandler struct {
	Store         *repository.Store
	WebAppURL     string
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	events, err := h.Store.ListEvents(r.Context(), middleware.OrganizerID(r))
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not list events")
		return
	}
	if events == nil {
		events = []models.Event{}
	}
	JSON(w, http.StatusOK, events)
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, http.StatusBadRequest, "title required")
		return
	}
	e, err := h.Store.CreateEvent(r.Context(), middleware.OrganizerID(r), req.Title)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create event")
		return
	}
	JSON(w, http.StatusCreated, map[string]any{
		"event":    e,
		"join_url": h.joinURL(e.JoinCode),
	})
}

func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ok, err := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	e, err := h.Store.GetEvent(r.Context(), eventID)
	if err != nil {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	count, _ := h.Store.AttendeeCount(r.Context(), eventID, 2*time.Minute)
	JSON(w, http.StatusOK, map[string]any{
		"event":          e,
		"join_url":       h.joinURL(e.JoinCode),
		"attendee_count": count,
	})
}

func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	var req struct {
		Title  *string `json:"title"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Status != nil {
		switch *req.Status {
		case "draft", "live", "ended":
		default:
			Error(w, http.StatusBadRequest, "status must be draft, live, or ended")
			return
		}
	}
	e, err := h.Store.UpdateEvent(r.Context(), eventID, req.Title, req.Status)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not update event")
		return
	}
	JSON(w, http.StatusOK, e)
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err := h.Store.DeleteEvent(r.Context(), eventID); err != nil {
		Error(w, http.StatusInternalServerError, "could not delete event")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandler) QR(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	e, err := h.Store.GetEvent(r.Context(), eventID)
	if err != nil {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	JSON(w, http.StatusOK, map[string]string{
		"join_url":  h.joinURL(e.JoinCode),
		"join_code": e.JoinCode,
	})
}

func (h *EventHandler) joinURL(code string) string {
	return h.WebAppURL + "/join/" + code
}
