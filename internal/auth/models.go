package auth

// User represents a user in the system.
type User struct {
	ID           int    `json:"id"`            // Unique identifier for the user
	UserUUID     string `json:"user_uuid"`     // UUID of the user
	Username     string `json:"username"`      // Username of the user
	PasswordHash string `json:"password_hash"` // Hashed password of the user
	CreatedAt    string `json:"created_at"`    // Timestamp of when the user was created
	UpdatedAt    string `json:"updated_at"`    // Timestamp of the last update to the user record
}

// Session represents a user session.
type Session struct {
	ID        int    `json:"id"`         // Unique identifier for the session
	SessionID string `json:"session_id"` // Session ID
	UserUUID  string `json:"user_uuid"`  // UUID of the user associated with the session
	ExpiresAt string `json:"expires_at"` // Timestamp of when the session expires
	CreatedAt string `json:"created_at"` // Timestamp of when the session was created
}

// RefreshToken represents a refresh token for a session.
type RefreshToken struct {
	ID        int    `json:"id"`         // Unique identifier for the refresh token
	Token     string `json:"token"`      // Refresh token string
	SessionID string `json:"session_id"` // Session ID associated with the refresh token
	CreatedAt string `json:"created_at"` // Timestamp of when the refresh token was created
}
