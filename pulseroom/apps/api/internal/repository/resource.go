package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pulseroom/api/internal/models"
)

func (s *Store) ListResources(ctx context.Context, eventID uuid.UUID) ([]models.Resource, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, event_id, title, url, file_key, kind, sort_order, created_at
		FROM resources WHERE event_id = $1 ORDER BY sort_order, created_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Resource
	for rows.Next() {
		var r models.Resource
		if err := rows.Scan(&r.ID, &r.EventID, &r.Title, &r.URL, &r.FileKey, &r.Kind, &r.SortOrder, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) CreateResource(ctx context.Context, eventID uuid.UUID, title, kind string, url *string) (*models.Resource, error) {
	var r models.Resource
	err := s.pool.QueryRow(ctx, `
		INSERT INTO resources (event_id, title, url, kind)
		VALUES ($1, $2, $3, $4)
		RETURNING id, event_id, title, url, file_key, kind, sort_order, created_at
	`, eventID, title, url, kind).Scan(
		&r.ID, &r.EventID, &r.Title, &r.URL, &r.FileKey, &r.Kind, &r.SortOrder, &r.CreatedAt,
	)
	return &r, err
}

func (s *Store) DeleteResource(ctx context.Context, eventID, resourceID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM resources WHERE id = $1 AND event_id = $2`, resourceID, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListAgenda(ctx context.Context, eventID uuid.UUID) ([]models.AgendaItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, event_id, title, speaker, starts_at, duration_minutes, sort_order
		FROM agenda_items WHERE event_id = $1 ORDER BY sort_order, starts_at NULLS LAST
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.AgendaItem
	for rows.Next() {
		var a models.AgendaItem
		if err := rows.Scan(&a.ID, &a.EventID, &a.Title, &a.Speaker, &a.StartsAt, &a.DurationMinutes, &a.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) CreateAgendaItem(ctx context.Context, eventID uuid.UUID, title string, speaker *string) (*models.AgendaItem, error) {
	var a models.AgendaItem
	err := s.pool.QueryRow(ctx, `
		INSERT INTO agenda_items (event_id, title, speaker)
		VALUES ($1, $2, $3)
		RETURNING id, event_id, title, speaker, starts_at, duration_minutes, sort_order
	`, eventID, title, speaker).Scan(
		&a.ID, &a.EventID, &a.Title, &a.Speaker, &a.StartsAt, &a.DurationMinutes, &a.SortOrder,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return &a, err
}

func (s *Store) DeleteAgendaItem(ctx context.Context, eventID, itemID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM agenda_items WHERE id = $1 AND event_id = $2`, itemID, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
