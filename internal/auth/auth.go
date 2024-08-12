package auth

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	DB              *database.Queries
	JWTSecretKey    string
	JWTRefreshKey   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Claims struct {
	UserUUID string `json:"user_uuid"`
	jwt.StandardClaims
}

// VerifyUsername checks if a username exists in the database.
// Returns true if the username does not exist, otherwise false.
func (auth *Authenticator) VerifyUsername(username string) (bool, error) {
	exists, err := auth.DB.VerifyUsername(context.Background(), username)
	if err != nil {
		log.WithError(err).Error("Failed to verify username")
		return false, err
	}
	return !exists, nil
}

// LoginUser authenticates the user and generates access and refresh tokens.
// Returns access token, refresh token, and any error encountered.
func (auth *Authenticator) LoginUser(username, password string) (string, string, error) {
	user, err := auth.DB.GetUserByUsername(context.Background(), username)
	if err != nil {
		log.WithError(err).Warn("Invalid username or password")
		return "", "", errors.New("invalid username or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.WithError(err).Warn("Invalid username or password")
		return "", "", errors.New("invalid username or password")
	}

	sessionID := uuid.New()
	expiresAt := time.Now().Add(auth.RefreshTokenTTL)

	// Create session in the database
	nullUserID := uuid.NullUUID{UUID: user.ID, Valid: true}
	if err := auth.DB.CreateSession(context.Background(), database.CreateSessionParams{
		SessionID: sessionID,
		UserID:    nullUserID,
		ExpiresAt: expiresAt,
	}); err != nil {
		log.WithError(err).Error("Failed to create session")
		return "", "", err
	}

	// Generate JWT tokens
	accessToken, err := auth.generateJWT(user.ID.String(), auth.AccessTokenTTL, auth.JWTSecretKey)
	if err != nil {
		log.WithError(err).Error("Failed to generate access token")
		return "", "", err
	}

	refreshToken, err := auth.generateJWT(sessionID.String(), auth.RefreshTokenTTL, auth.JWTRefreshKey)
	if err != nil {
		log.WithError(err).Error("Failed to generate refresh token")
		return "", "", err
	}

	// Store refresh token in the database
	if err := auth.DB.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		ID:    sessionID,
	}); err != nil {
		log.WithError(err).Error("Failed to store refresh token")
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshToken generates a new access token using a valid refresh token.
// Returns the new access token and any error encountered.
func (auth *Authenticator) RefreshToken(refreshToken string) (string, error) {
	// Parse and validate the refresh token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.JWTRefreshKey), nil
	})
	if err != nil || !token.Valid {
		log.WithError(err).Warn("Invalid refresh token")
		return "", errors.New("invalid refresh token")
	}

	// Retrieve session and user information
	sessionID, err := auth.DB.GetSessionIDByRefreshToken(context.Background(), refreshToken)
	if err != nil {
		log.WithError(err).Warn("Invalid refresh token")
		return "", errors.New("invalid refresh token")
	}

	userUUID, err := auth.DB.GetUserUUIDBySessionID(context.Background(), sessionID)
	if err != nil {
		log.WithError(err).Warn("Invalid session")
		return "", errors.New("invalid session")
	}

	// Generate new access token
	newAccessToken, err := auth.generateJWT(userUUID.UUID.String(), auth.AccessTokenTTL, auth.JWTSecretKey)
	if err != nil {
		log.WithError(err).Error("Failed to generate new access token")
		return "", err
	}

	return newAccessToken, nil
}

// Logout removes the session associated with the provided session ID.
// Returns any error encountered during the operation.
func (auth *Authenticator) Logout(sessionID string) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		log.WithError(err).Error("Error parsing UUID")
		return err
	}

	// Delete session from the database
	if err := auth.DB.DeleteSession(context.Background(), sessionUUID); err != nil {
		log.WithError(err).Error("Failed to logout")
		return err
	}

	return nil
}

// Authenticate verifies the provided JWT token and returns the user UUID if valid.
// Returns the user UUID and any error encountered.
func (auth *Authenticator) Authenticate(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		log.WithError(err).Warn("Invalid token")
		return "", errors.New("invalid token")
	}

	return claims.UserUUID, nil
}

// generateJWT creates a JWT token with the specified subject, TTL, and secret key.
// Returns the signed JWT and any error encountered.
func (auth *Authenticator) generateJWT(subject string, ttl time.Duration, secretKey string) (string, error) {
	claims := &Claims{
		UserUUID: subject,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
		},
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.WithError(err).Error("Failed to sign JWT")
		return "", err
	}

	return signedToken, nil
}
