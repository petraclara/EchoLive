package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pulseroom/api/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateOrganizer(ctx context.Context, email, passwordHash, name string) (*models.Organizer, error) {
	var o models.Organizer
	err := s.pool.QueryRow(ctx, `
		INSERT INTO organizers (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, created_at
	`, email, passwordHash, name).Scan(&o.ID, &o.Email, &o.Name, &o.CreatedAt)
	return &o, err
}

func (s *Store) GetOrganizerByEmail(ctx context.Context, email string) (id uuid.UUID, passwordHash, name string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT id, password_hash, name FROM organizers WHERE email = $1
	`, email).Scan(&id, &passwordHash, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, "", "", ErrNotFound
	}
	return
}

func (s *Store) GetOrganizer(ctx context.Context, id uuid.UUID) (*models.Organizer, error) {
	var o models.Organizer
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, created_at FROM organizers WHERE id = $1
	`, id).Scan(&o.ID, &o.Email, &o.Name, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &o, err
}
