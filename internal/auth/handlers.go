package auth

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type UserHandler struct {
	Authenticator *Authenticator
}

// VerifyUsername checks if a username is available.
// It responds with a 200 OK if the username is available
func (handler *UserHandler) VerifyUsername(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}

	// Decode request payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if the username is available
	available, err := handler.Authenticator.VerifyUsername(req.Username)
	if err != nil {
		log.WithError(err).Error("Error verifying username")
		http.Error(w, "Error verifying username", http.StatusInternalServerError)
		return
	}

	// Respond based on availability
	if !available {
		log.Infof("Username %s is not available", req.Username)
		http.Error(w, "Username not available", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Infof("Username %s is available", req.Username)
}

// RefreshToken generates a new access token using the provided refresh token.
// It responds with a 200 OK and the new access token,
// or 400 Bad Request if the payload is invalid,
// 401 Unauthorized if the refresh token is invalid
func (handler *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Decode request payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate new access token
	newAccessToken, err := handler.Authenticator.RefreshToken(req.RefreshToken)
	if err != nil {
		log.WithError(err).Warn("Invalid refresh token")
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Respond with the new access token
	resp := map[string]string{
		"access_token": newAccessToken,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.WithError(err).Error("Failed to encode response")
		http.Error(w, "Failed to generate token response", http.StatusInternalServerError)
		return
	}

	log.Info("Token refreshed successfully")
}

// Logout removes the session associated with the provided session ID.
// It responds with a 200 OK if the session is successfully logged out,
// or 400 Bad Request if the session ID is missing.
func (handler *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		log.Warn("Session ID required for logout")
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Log out the session
	if err := handler.Authenticator.Logout(sessionID); err != nil {
		log.WithError(err).Error("Error logging out")
		http.Error(w, "Error logging out", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Infof("Session %s logged out successfully", sessionID)
}
