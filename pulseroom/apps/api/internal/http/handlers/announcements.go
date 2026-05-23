package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/http/middleware"
	"github.com/pulseroom/api/internal/repository"
	"github.com/pulseroom/api/internal/ws"
)

type AnnouncementHandler struct {
	Store *repository.Store
	Hub   *ws.Hub
}

func (h *AnnouncementHandler) broadcast(eventID uuid.UUID, msgType string, payload any) {
	b, _ := json.Marshal(payload)
	h.Hub.Broadcast(eventID, ws.Message{
		Type:    msgType,
		EventID: eventID,
		Payload: b,
		TS:      time.Now().Unix(),
	})
}

func (h *AnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	list, err := h.Store.ListAnnouncements(r.Context(), eventID, nil, 100)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not list announcements")
		return
	}
	JSON(w, http.StatusOK, list)
}

func (h *AnnouncementHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		Body    string  `json:"body"`
		Type    string  `json:"type"`
		LinkURL *string `json:"link_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" {
		Error(w, http.StatusBadRequest, "body required")
		return
	}
	if req.Type == "" {
		req.Type = "info"
	}
	a, err := h.Store.CreateAnnouncement(r.Context(), eventID, req.Body, req.Type, req.LinkURL)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create announcement")
		return
	}
	h.broadcast(eventID, "announcement.created", a)
	JSON(w, http.StatusCreated, a)
}

func (h *AnnouncementHandler) Pin(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	aid, err := uuid.Parse(chi.URLParam(r, "announcementID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	a, err := h.Store.PinAnnouncement(r.Context(), eventID, aid)
	if err == repository.ErrNotFound {
		Error(w, http.StatusNotFound, "announcement not found")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not pin")
		return
	}
	h.broadcast(eventID, "announcement.pinned", a)
	JSON(w, http.StatusOK, a)
}

func (h *AnnouncementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	aid, err := uuid.Parse(chi.URLParam(r, "announcementID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err := h.Store.DeleteAnnouncement(r.Context(), eventID, aid); err != nil {
		Error(w, http.StatusNotFound, "announcement not found")
		return
	}
	h.broadcast(eventID, "announcement.deleted", map[string]string{"id": aid.String()})
	w.WriteHeader(http.StatusNoContent)
}
