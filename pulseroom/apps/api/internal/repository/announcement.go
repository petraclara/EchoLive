package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pulseroom/api/internal/models"
)

func (s *Store) CreateAnnouncement(ctx context.Context, eventID uuid.UUID, body, annType string, linkURL *string) (*models.Announcement, error) {
	var a models.Announcement
	err := s.pool.QueryRow(ctx, `
		INSERT INTO announcements (event_id, body, type, link_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, event_id, body, type, link_url, is_pinned, created_at
	`, eventID, body, annType, linkURL).Scan(
		&a.ID, &a.EventID, &a.Body, &a.Type, &a.LinkURL, &a.IsPinned, &a.CreatedAt,
	)
	return &a, err
}

func (s *Store) ListAnnouncements(ctx context.Context, eventID uuid.UUID, since *time.Time, limit int) ([]models.Announcement, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if since != nil {
		rows, err = s.pool.Query(ctx, `
			SELECT id, event_id, body, type, link_url, is_pinned, created_at
			FROM announcements WHERE event_id = $1 AND created_at > $2
			ORDER BY created_at ASC LIMIT $3
		`, eventID, since, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, event_id, body, type, link_url, is_pinned, created_at
			FROM announcements WHERE event_id = $1
			ORDER BY created_at DESC LIMIT $2
		`, eventID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Announcement
	for rows.Next() {
		var a models.Announcement
		if err := rows.Scan(&a.ID, &a.EventID, &a.Body, &a.Type, &a.LinkURL, &a.IsPinned, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	if since == nil {
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}
	return list, rows.Err()
}

func (s *Store) GetPinnedAnnouncement(ctx context.Context, eventID uuid.UUID) (*models.Announcement, error) {
	var a models.Announcement
	err := s.pool.QueryRow(ctx, `
		SELECT id, event_id, body, type, link_url, is_pinned, created_at
		FROM announcements WHERE event_id = $1 AND is_pinned = true LIMIT 1
	`, eventID).Scan(&a.ID, &a.EventID, &a.Body, &a.Type, &a.LinkURL, &a.IsPinned, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &a, err
}

func (s *Store) PinAnnouncement(ctx context.Context, eventID, announcementID uuid.UUID) (*models.Announcement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE announcements SET is_pinned = false WHERE event_id = $1`, eventID)
	if err != nil {
		return nil, err
	}
	var a models.Announcement
	err = tx.QueryRow(ctx, `
		UPDATE announcements SET is_pinned = true
		WHERE id = $1 AND event_id = $2
		RETURNING id, event_id, body, type, link_url, is_pinned, created_at
	`, announcementID, eventID).Scan(
		&a.ID, &a.EventID, &a.Body, &a.Type, &a.LinkURL, &a.IsPinned, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, tx.Commit(ctx)
}

func (s *Store) DeleteAnnouncement(ctx context.Context, eventID, announcementID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM announcements WHERE id = $1 AND event_id = $2`, announcementID, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
