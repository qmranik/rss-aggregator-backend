package auth

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	log "github.com/sirupsen/logrus"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

// MiddlewareAuth is a middleware function that authenticates a user based on the Authorization header.
// It extracts the token, validates it, retrieves the user from the database, and then calls the provided handler.
func (auth *Authenticator) MiddlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.WithFields(log.Fields{
				"func":  "MiddlewareAuth",
				"event": "AuthorizationHeaderMissing",
			}).Error("Authorization header missing")
			helper.RespondWithError(w, http.StatusUnauthorized, "Authorization header missing")
			return
		}

		// Split "Bearer token" from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // Authorization header doesn't have "Bearer " prefix
			log.WithFields(log.Fields{
				"func":  "MiddlewareAuth",
				"event": "InvalidAuthorizationFormat",
			}).Error("Invalid authorization header format")
			helper.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		// Validate token and get user UUID
		userUUID, err := auth.Authenticate(tokenString)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"func":  "MiddlewareAuth",
				"event": "Authenticate",
			}).Error("Invalid token")
			helper.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		uuid, err := uuid.Parse(userUUID)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"func":  "MiddlewareAuth",
				"event": "ParseUserUUID",
			}).Error("Couldn't parse userUUID to uuid")
			helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't parse user UUID")
			return
		}

		// Retrieve user from database
		user, err := auth.DB.GetUserByID(r.Context(), uuid)
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err,
				"func":     "MiddlewareAuth",
				"event":    "GetUserByID",
				"userUUID": userUUID,
			}).Error("Couldn't get user")
			helper.RespondWithError(w, http.StatusNotFound, "Couldn't get user")
			return
		}

		// Call the handler with the authenticated user
		handler(w, r, user)
	}
}
