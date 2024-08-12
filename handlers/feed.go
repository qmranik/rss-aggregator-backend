package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

// HandlerFeedCreate creates a new feed and automatically follows it for the user.
func (cfg *ApiConfig) HandlerFeedCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	// Decode the incoming request body into the parameters struct
	var params models.Parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerFeedCreate",
		}).Error("Couldn't decode parameters")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// Create a new feed record in the database
	feed, err := cfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		Name:      params.Name,
		Url:       params.URL,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"func":     "HandlerFeedCreate",
			"userID":   user.ID,
			"feedName": params.Name,
			"feedURL":  params.URL,
		}).Error("Couldn't create feed")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create feed")
		return
	}

	// Automatically follow the newly created feed
	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerFeedCreate",
			"userID": user.ID,
			"feedID": feed.ID,
		}).Error("Couldn't create feed follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create feed follow")
		return
	}

	// Respond with the created feed and feed follow details
	helper.RespondWithJSON(w, http.StatusOK, struct {
		Feed       models.Feed       `json:"feed"`
		FeedFollow models.FeedFollow `json:"feed_follow"`
	}{
		Feed:       models.DatabaseFeedToFeed(feed),
		FeedFollow: models.DatabaseFeedFollowToFeedFollow(feedFollow),
	})
}

// HandlerGetFeeds retrieves all feeds from the database.
func (cfg *ApiConfig) HandlerGetFeeds(w http.ResponseWriter, r *http.Request) {
	// Fetch all feeds from the database
	feeds, err := cfg.DB.GetFeeds(r.Context())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerGetFeeds",
		}).Error("Couldn't get feeds")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't get feeds")
		return
	}

	// Respond with the retrieved feeds in JSON format
	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseFeedsToFeeds(feeds))
}
