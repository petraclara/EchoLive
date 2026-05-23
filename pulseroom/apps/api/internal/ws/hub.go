package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

type Message struct {
	Type    string          `json:"type"`
	EventID uuid.UUID       `json:"event_id"`
	Payload json.RawMessage `json:"payload"`
	TS      int64           `json:"ts"`
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]*Room
}

type Room struct {
	eventID uuid.UUID
	clients map[*Client]bool
	broadcast chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uuid.UUID]*Room)}
}

func (h *Hub) Room(eventID uuid.UUID) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[eventID]; ok {
		return r
	}
	r := &Room{
		eventID:    eventID,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	h.rooms[eventID] = r
	go r.run()
	return r
}

func (h *Hub) Broadcast(eventID uuid.UUID, msg Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.Room(eventID).broadcast <- b
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.register:
			r.clients[c] = true
		case c := <-r.unregister:
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
				close(c.Send)
			}
		case msg := <-r.broadcast:
			for c := range r.clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
					delete(r.clients, c)
				}
			}
		}
	}
}

func (r *Room) Register(c *Client) {
	r.register <- c
}

func (r *Room) Unregister(c *Client) {
	r.unregister <- c
}

func (h *Hub) ClientCount(eventID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[eventID]
	if !ok {
		return 0
	}
	return len(r.clients)
}
