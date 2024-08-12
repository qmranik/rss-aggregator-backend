package handlers

import (
	"net/http"
	"strconv"

	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

// HandlerPostsGet retrieves posts for the authenticated user with an optional limit parameter.
func (cfg *ApiConfig) HandlerPostsGet(w http.ResponseWriter, r *http.Request, user database.User) {
	// Default limit for posts returned
	limit := 10

	// Attempt to retrieve the limit from query parameters, fallback to default if invalid
	limitStr := r.URL.Query().Get("limit")
	if specifiedLimit, err := strconv.Atoi(limitStr); err == nil {
		limit = specifiedLimit
	} else if limitStr != "" {
		// Log a warning if the limit parameter was provided but invalid
		log.WithFields(log.Fields{
			"error":    err,
			"func":     "HandlerPostsGet",
			"limitStr": limitStr,
		}).Warn("Invalid limit parameter, using default limit")
	}

	// Fetch posts for the user from the database, limited by the specified or default limit
	posts, err := cfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		// Log an error and respond with a 500 status code if the database query fails
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerPostsGet",
			"userID": user.ID,
			"limit":  limit,
		}).Error("Couldn't get posts for user")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't get posts for user")
		return
	}

	// Respond with the retrieved posts in JSON format
	helper.RespondWithJSON(w, http.StatusOK, models.DatabasePostsToPosts(posts))
}
