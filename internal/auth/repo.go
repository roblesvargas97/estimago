package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertUser persists a new user with the provided password hash.
func InsertUser(ctx context.Context, pool *pgxpool.Pool, name, email, passwordHash string) (User, error) {
	const planID = "free"

	var u User
	err := pool.QueryRow(ctx, `
                INSERT INTO users (name, email, password_hash, plan_id)
                VALUES ($1, $2, $3, $4)
                RETURNING id, name, email, password_hash, plan_id, created_at, updated_at
        `, name, email, passwordHash, planID).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}

// GetUserByEmail fetches a user by email.
func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (User, error) {
	var u User
	err := pool.QueryRow(ctx, `
                SELECT id, name, email, password_hash, plan_id, created_at, updated_at
                FROM users
                WHERE email = $1
        `, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}

// GetUserByID fetches a user by its ID.
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (User, error) {
	var u User
	err := pool.QueryRow(ctx, `
                SELECT id, name, email, password_hash, plan_id, created_at, updated_at
                FROM users
                WHERE id = $1
        `, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}

// IsNotFound reports whether the error is pgx.ErrNoRows.
func IsNotFound(err error) bool {
	return err != nil && err == pgx.ErrNoRows
}
