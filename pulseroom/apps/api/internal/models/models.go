package models

import (
	"time"

	"github.com/google/uuid"
)

type Organizer struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Event struct {
	ID          uuid.UUID  `json:"id"`
	OrganizerID uuid.UUID  `json:"organizer_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	JoinCode    string     `json:"join_code"`
	Status      string     `json:"status"`
	StartsAt    *time.Time `json:"starts_at,omitempty"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Announcement struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	LinkURL   *string   `json:"link_url,omitempty"`
	IsPinned  bool      `json:"is_pinned"`
	CreatedAt time.Time `json:"created_at"`
}

type Resource struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	Title     string    `json:"title"`
	URL       *string   `json:"url,omitempty"`
	FileKey   *string   `json:"file_key,omitempty"`
	Kind      string    `json:"kind"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type AgendaItem struct {
	ID              uuid.UUID  `json:"id"`
	EventID         uuid.UUID  `json:"event_id"`
	Title           string     `json:"title"`
	Speaker         *string    `json:"speaker,omitempty"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	SortOrder       int        `json:"sort_order"`
}

type AttendeeSession struct {
	ID           uuid.UUID `json:"id"`
	EventID      uuid.UUID `json:"event_id"`
	SessionToken string    `json:"-"`
}

type PublicEvent struct {
	Event
	Pinned         *Announcement  `json:"pinned,omitempty"`
	Announcements  []Announcement `json:"announcements"`
	Resources      []Resource     `json:"resources"`
	Agenda         []AgendaItem   `json:"agenda"`
	AttendeeCount  int            `json:"attendee_count"`
}
