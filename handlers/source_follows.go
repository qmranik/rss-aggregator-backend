package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

func (cfg *ApiConfig) HandlerSourceFollowsGet(w http.ResponseWriter, r *http.Request, user database.User) {
	sourceFollows, err := cfg.DB.GetSourceFollowsForUser(r.Context(), user.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerSourceFollowsGet",
			"userID": user.ID,
		}).Error("Couldn't get source follows for user")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't retrieve source follows")
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseSourceFollowsToSourceFollows(sourceFollows))
}

func (cfg *ApiConfig) HandlerSourceFollowCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		SourceID uuid.UUID `json:"source_id"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerSourceFollowCreate",
		}).Error("Couldn't decode parameters")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	sourceFollow, err := cfg.DB.CreateSourceFollow(r.Context(), database.CreateSourceFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		SourceID:  params.SourceID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"func":     "HandlerSourceFollowCreate",
			"userID":   user.ID,
			"sourceID": params.SourceID,
		}).Error("Couldn't create source follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create source follow")
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseSourceFollowToSourceFollow(sourceFollow))
}

func (cfg *ApiConfig) HandlerSourceFollowDelete(w http.ResponseWriter, r *http.Request, user database.User) {
	sourceFollowIDStr := chi.URLParam(r, "sourceFollowID")
	sourceFollowID, err := uuid.Parse(sourceFollowIDStr)
	if err != nil {
		log.WithFields(log.Fields{
			"error":             err,
			"func":              "HandlerSourceFollowDelete",
			"sourceFollowIDStr": sourceFollowIDStr,
		}).Error("Invalid source follow ID")
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid source follow ID")
		return
	}

	err = cfg.DB.DeleteSourceFollow(r.Context(), database.DeleteSourceFollowParams{
		UserID: user.ID,
		ID:     sourceFollowID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":          err,
			"func":           "HandlerSourceFollowDelete",
			"userID":         user.ID,
			"sourceFollowID": sourceFollowID,
		}).Error("Couldn't delete source follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete source follow")
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, struct{}{})
}
