package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"regexp"
	"strings"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pulseroom/api/internal/models"
)

var joinChars = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

func generateJoinCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = joinChars[int(b[i])%len(joinChars)]
	}
	return string(b), nil
}

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "event"
	}
	return s
}

func (s *Store) CreateEvent(ctx context.Context, organizerID uuid.UUID, title string) (*models.Event, error) {
	code, err := generateJoinCode()
	if err != nil {
		return nil, err
	}
	slug := slugify(title)
	var e models.Event
	err = s.pool.QueryRow(ctx, `
		INSERT INTO events (organizer_id, title, slug, join_code)
		VALUES ($1, $2, $3, $4)
		RETURNING id, organizer_id, title, slug, join_code, status, starts_at, ends_at, created_at, updated_at
	`, organizerID, title, slug, code).Scan(
		&e.ID, &e.OrganizerID, &e.Title, &e.Slug, &e.JoinCode, &e.Status,
		&e.StartsAt, &e.EndsAt, &e.CreatedAt, &e.UpdatedAt,
	)
	return &e, err
}

func (s *Store) ListEvents(ctx context.Context, organizerID uuid.UUID) ([]models.Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organizer_id, title, slug, join_code, status, starts_at, ends_at, created_at, updated_at
		FROM events WHERE organizer_id = $1 ORDER BY created_at DESC
	`, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.OrganizerID, &e.Title, &e.Slug, &e.JoinCode, &e.Status,
			&e.StartsAt, &e.EndsAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *Store) GetEvent(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var e models.Event
	err := s.pool.QueryRow(ctx, `
		SELECT id, organizer_id, title, slug, join_code, status, starts_at, ends_at, created_at, updated_at
		FROM events WHERE id = $1
	`, id).Scan(&e.ID, &e.OrganizerID, &e.Title, &e.Slug, &e.JoinCode, &e.Status,
		&e.StartsAt, &e.EndsAt, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &e, err
}

func (s *Store) GetEventByJoinCode(ctx context.Context, code string) (*models.Event, error) {
	var e models.Event
	err := s.pool.QueryRow(ctx, `
		SELECT id, organizer_id, title, slug, join_code, status, starts_at, ends_at, created_at, updated_at
		FROM events WHERE join_code = $1
	`, strings.ToUpper(strings.TrimSpace(code))).Scan(
		&e.ID, &e.OrganizerID, &e.Title, &e.Slug, &e.JoinCode, &e.Status,
		&e.StartsAt, &e.EndsAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &e, err
}

func (s *Store) UpdateEvent(ctx context.Context, id uuid.UUID, title, status *string) (*models.Event, error) {
	e, err := s.GetEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	if title != nil {
		e.Title = *title
		e.Slug = slugify(*title)
	}
	if status != nil {
		e.Status = *status
	}
	err = s.pool.QueryRow(ctx, `
		UPDATE events SET title = $2, slug = $3, status = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, organizer_id, title, slug, join_code, status, starts_at, ends_at, created_at, updated_at
	`, id, e.Title, e.Slug, e.Status).Scan(
		&e.ID, &e.OrganizerID, &e.Title, &e.Slug, &e.JoinCode, &e.Status,
		&e.StartsAt, &e.EndsAt, &e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

func (s *Store) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	return err
}

func (s *Store) EventOwnedBy(ctx context.Context, eventID, organizerID uuid.UUID) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND organizer_id = $2)
	`, eventID, organizerID).Scan(&ok)
	return ok, err
}
