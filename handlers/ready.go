package handlers

import (
	"net/http"

	"github.com/qmranik/rss-aggregator-backend/helper"
)

// HandlerReadiness provides a simple health check endpoint to verify if the service is ready.
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	helper.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
