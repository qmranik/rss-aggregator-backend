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

func (cfg *ApiConfig) HandlerSourceCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	var params models.SourceParameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerSourceCreate",
		}).Error("Couldn't decode parameters")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	sourceType := database.SourceTypeEnum(params.SourceType)
	if sourceType == "" {
		sourceType = database.SourceTypeEnumRss
	}
	
	fetchInterval := database.FetchIntervalEnum(params.FetchInterval)
	if fetchInterval == "" {
		fetchInterval = database.FetchIntervalEnumHourly
	}

	source, err := cfg.DB.CreateSource(r.Context(), database.CreateSourceParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		Name:      params.Name,
		Url:       params.URL,
		SourceType: sourceType,
		FetchInterval: fetchInterval,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"func":     "HandlerSourceCreate",
			"userID":   user.ID,
		}).Error("Couldn't create source")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create source")
		return
	}

	sourceFollow, err := cfg.DB.CreateSourceFollow(r.Context(), database.CreateSourceFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		SourceID:  source.ID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"func":   "HandlerSourceCreate",
			"userID": user.ID,
		}).Error("Couldn't create source follow")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create source follow")
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, struct {
		Source       models.Source       `json:"source"`
		SourceFollow models.SourceFollow `json:"source_follow"`
	}{
		Source:       models.DatabaseSourceToSource(source),
		SourceFollow: models.DatabaseSourceFollowToSourceFollow(sourceFollow),
	})
}

func (cfg *ApiConfig) HandlerGetSources(w http.ResponseWriter, r *http.Request) {
	sources, err := cfg.DB.GetSources(r.Context())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerGetSources",
		}).Error("Couldn't get sources")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't get sources")
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseSourcesToSources(sources))
}
