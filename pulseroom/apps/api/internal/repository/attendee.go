package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/models"
)

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) CreateAttendeeSession(ctx context.Context, eventID uuid.UUID, userAgent string) (*models.AttendeeSession, string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, "", err
	}
	var sess models.AttendeeSession
	err = s.pool.QueryRow(ctx, `
		INSERT INTO attendee_sessions (event_id, session_token, user_agent)
		VALUES ($1, $2, $3)
		RETURNING id, event_id
	`, eventID, token, userAgent).Scan(&sess.ID, &sess.EventID)
	return &sess, token, err
}

func (s *Store) ValidateAttendeeSession(ctx context.Context, eventID uuid.UUID, token string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM attendee_sessions WHERE event_id = $1 AND session_token = $2)
	`, eventID, token).Scan(&ok)
	return ok, err
}

func (s *Store) TouchAttendeeSession(ctx context.Context, eventID uuid.UUID, token string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE attendee_sessions SET last_seen_at = now()
		WHERE event_id = $1 AND session_token = $2
	`, eventID, token)
	return err
}

func (s *Store) AttendeeCount(ctx context.Context, eventID uuid.UUID, _ time.Duration) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM attendee_sessions
		WHERE event_id = $1 AND last_seen_at > now() - interval '2 minutes'
	`, eventID).Scan(&count)
	return count, err
}
