package helper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/araddon/dateparse"
	log "github.com/sirupsen/logrus"
)

// RespondWithError sends an error response with the given HTTP status code and message.
// It also logs the error if the status code indicates a server error (5XX).
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if code >= 500 {
		log.WithFields(log.Fields{
			"code":  code,
			"error": msg,
		}).Error("Responding with 5XX error")
	}

	// Create an error response structure
	type errorResponse struct {
		Error string `json:"error"`
	}

	// Respond with JSON formatted error
	RespondWithJSON(w, code, errorResponse{Error: msg})
}

// HandlerReadiness performs a readiness check and responds with a status of "ok".
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandlerErr handles generic internal server errors by sending a 500 response.
func HandlerErr(w http.ResponseWriter, r *http.Request) {
	RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}

// RespondWithJSON sends a JSON response with the given HTTP status code and payload.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Marshal the payload into JSON
	dat, err := json.Marshal(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"payload": payload,
		}).Error("Error marshalling JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the response code and write the JSON data
	w.WriteHeader(code)
	w.Write(dat)
}

// ParsePubDate parses a pubDate string into a sql.NullTime, returning an error if parsing fails.
func ParsePubDate(pubDate string) (sql.NullTime, error) {
	t, err := dateparse.ParseAny(pubDate)
	if err != nil {
		log.WithFields(log.Fields{
			"pubDate": pubDate,
			"error":   err,
		}).Error("Unable to parse pubDate")
		return sql.NullTime{}, fmt.Errorf("unable to parse pubDate: %w", err)
	}

	// Return the parsed time as sql.NullTime
	return sql.NullTime{
		Time:  t,
		Valid: true,
	}, nil
}
