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

type ResourceHandler struct {
	Store *repository.Store
	Hub   *ws.Hub
}

func (h *ResourceHandler) List(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	list, err := h.Store.ListResources(r.Context(), eventID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not list resources")
		return
	}
	JSON(w, http.StatusOK, list)
}

func (h *ResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	var req struct {
		Title string  `json:"title"`
		URL   *string `json:"url"`
		Kind  string  `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, http.StatusBadRequest, "title required")
		return
	}
	if req.Kind == "" {
		req.Kind = "link"
	}
	res, err := h.Store.CreateResource(r.Context(), eventID, req.Title, req.Kind, req.URL)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create resource")
		return
	}
	h.notify(eventID, "resource.updated", res)
	JSON(w, http.StatusCreated, res)
}

func (h *ResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	rid, _ := uuid.Parse(chi.URLParam(r, "resourceID"))
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err := h.Store.DeleteResource(r.Context(), eventID, rid); err != nil {
		Error(w, http.StatusNotFound, "resource not found")
		return
	}
	h.notify(eventID, "resource.deleted", map[string]string{"id": rid.String()})
	w.WriteHeader(http.StatusNoContent)
}

func (h *ResourceHandler) ListAgenda(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	list, err := h.Store.ListAgenda(r.Context(), eventID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not list agenda")
		return
	}
	JSON(w, http.StatusOK, list)
}

func (h *ResourceHandler) CreateAgenda(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	var req struct {
		Title   string  `json:"title"`
		Speaker *string `json:"speaker"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, http.StatusBadRequest, "title required")
		return
	}
	item, err := h.Store.CreateAgendaItem(r.Context(), eventID, req.Title, req.Speaker)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create agenda item")
		return
	}
	h.notify(eventID, "agenda.updated", item)
	JSON(w, http.StatusCreated, item)
}

func (h *ResourceHandler) DeleteAgenda(w http.ResponseWriter, r *http.Request) {
	eventID, _ := uuid.Parse(chi.URLParam(r, "eventID"))
	itemID, _ := uuid.Parse(chi.URLParam(r, "itemID"))
	ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, middleware.OrganizerID(r))
	if !ok {
		Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err := h.Store.DeleteAgendaItem(r.Context(), eventID, itemID); err != nil {
		Error(w, http.StatusNotFound, "item not found")
		return
	}
	h.notify(eventID, "agenda.deleted", map[string]string{"id": itemID.String()})
	w.WriteHeader(http.StatusNoContent)
}

func (h *ResourceHandler) notify(eventID uuid.UUID, msgType string, payload any) {
	b, _ := json.Marshal(payload)
	h.Hub.Broadcast(eventID, ws.Message{
		Type:    msgType,
		EventID: eventID,
		Payload: b,
		TS:      time.Now().Unix(),
	})
}
