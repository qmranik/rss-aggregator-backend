package handlers

import (
	"github.com/qmranik/rss-aggregator-backend/internal/auth"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
)

// ApiConfig contains the database and authentication configurations for the API.
type ApiConfig struct {
	DB   *database.Queries   // DB is a pointer to the database queries interface.
	Auth *auth.Authenticator // Auth is a pointer to the authentication manager.
}
