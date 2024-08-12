package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

// HandlerFeedFollowsGet retrieves all feed follows for a given user.
func (cfg *ApiConfig) HandlerFeedFollowsGet(w http.ResponseWriter, r *http.Request, user database.User) {
	// Fetch feed follows for the user from the database
	feedFollows, err := cfg.DB.GetFeedFollowsForUser(r.Context(), user.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerFeedFollowsGet",
			"userID": user.ID,
		}).Error("Couldn't get feed follows for user")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't retrieve feed follows")
		return
	}

	// Respond with the retrieved feed follows in JSON format
	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseFeedFollowsToFeedFollows(feedFollows))
}

// HandlerFeedFollowCreate creates a new feed follow for a given user.
func (cfg *ApiConfig) HandlerFeedFollowCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	// Define a struct to capture the incoming parameters
	type parameters struct {
		FeedID uuid.UUID
	}

	// Decode the request body into the parameters struct
	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerFeedFollowCreate",
		}).Error("Couldn't decode parameters")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// Create a new feed follow record in the database
	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    params.FeedID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerFeedFollowCreate",
			"userID": user.ID,
			"feedID": params.FeedID,
		}).Error("Couldn't create feed follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create feed follow")
		return
	}

	// Respond with the created feed follow in JSON format
	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseFeedFollowToFeedFollow(feedFollow))
}

// HandlerFeedFollowDelete deletes an existing feed follow for a given user.
func (cfg *ApiConfig) HandlerFeedFollowDelete(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract the feed follow ID from the URL parameters
	feedFollowIDStr := chi.URLParam(r, "feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err,
			"func":            "HandlerFeedFollowDelete",
			"feedFollowIDStr": feedFollowIDStr,
		}).Error("Invalid feed follow ID")
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}

	// Delete the feed follow record from the database
	err = cfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		ID:     feedFollowID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"func":         "HandlerFeedFollowDelete",
			"userID":       user.ID,
			"feedFollowID": feedFollowID,
		}).Error("Couldn't delete feed follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete feed follow")
		return
	}

	// Respond with an empty JSON object indicating success
	helper.RespondWithJSON(w, http.StatusOK, struct{}{})
}
