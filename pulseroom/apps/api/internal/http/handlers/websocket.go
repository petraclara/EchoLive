package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/auth"
	"github.com/pulseroom/api/internal/http/middleware"
	"github.com/pulseroom/api/internal/repository"
	"github.com/pulseroom/api/internal/ws"
)

type WSHandler struct {
	Store          *repository.Store
	Hub            *ws.Hub
	JWTSecret      string
	AllowedOrigins []string
}

func (h *WSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid event id")
		return
	}
	if _, err := h.Store.GetEvent(r.Context(), eventID); err != nil {
		Error(w, http.StatusNotFound, "event not found")
		return
	}

	token := middleware.BearerFromRequest(r)
	if token == "" {
		Error(w, http.StatusUnauthorized, "token required")
		return
	}

	isOrganizer := false
	if claims, err := auth.ParseToken(h.JWTSecret, token); err == nil {
		ok, _ := h.Store.EventOwnedBy(r.Context(), eventID, claims.OrganizerID)
		isOrganizer = ok
	}

	if !isOrganizer {
		ok, err := h.Store.ValidateAttendeeSession(r.Context(), eventID, token)
		if err != nil || !ok {
			Error(w, http.StatusUnauthorized, "invalid session")
			return
		}
	}

	upgrader := ws.NewUpgrader(h.AllowedOrigins)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	room := h.Hub.Room(eventID)
	client := &ws.Client{
		Hub:  room,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	room.Register(client)

	go client.WritePump()
	go client.ReadPump(func(msg []byte) {
		var envelope struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(msg, &envelope) == nil && envelope.Type == "presence.heartbeat" {
			_ = h.Store.TouchAttendeeSession(r.Context(), eventID, token)
		}
	})

	count, _ := h.Store.AttendeeCount(r.Context(), eventID, 2*time.Minute)
	payload, _ := json.Marshal(map[string]int{"count": count})
	h.Hub.Broadcast(eventID, ws.Message{
		Type:    "presence.count",
		EventID: eventID,
		Payload: payload,
		TS:      time.Now().Unix(),
	})
}
