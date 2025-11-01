package auth

import (
	"time"

	"github.com/google/uuid"
)

// User models a system user stored in the database.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	PlanID       string    `json:"plan_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SignupIn captures the signup payload.
type SignupIn struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupOut returns the public user data after signup.
type SignupOut struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	PlanID    string    `json:"plan_id"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginIn captures the login payload.
type LoginIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginOut contains the login response body.
type LoginOut struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

// AuthUser is the minimal user payload returned to clients.
type AuthUser struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	PlanID string    `json:"plan_id"`
}

// Config wires JWT configuration for handlers and middleware.
type Config struct {
	JWTSecret   string
	JWTTTLHours int
}
