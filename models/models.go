package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
)

// Parameters represents a set of parameters for feed creation or updates.
type Parameters struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// UserPram represents user credentials for login or registration.
type UserPram struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// User represents a user in the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Name         string    `json:"name"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

// DatabaseUserToUser converts a database.User to a User model.
func DatabaseUserToUser(user database.User) User {
	return User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Name:         user.Username,
		AccessToken:  "", // Assuming you handle tokens elsewhere
		RefreshToken: "", // Assuming you handle tokens elsewhere
	}
}

// NullTimeToTimePtr converts a sql.NullTime to a *time.Time pointer.
func NullTimeToTimePtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

// NullStringToStringPtr converts a sql.NullString to a *string pointer.
func NullStringToStringPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}
