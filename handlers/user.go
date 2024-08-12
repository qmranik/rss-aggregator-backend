package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// HandlerUsersCreate creates a new user and responds with authentication tokens.
func (cfg *ApiConfig) HandlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Decode the incoming request body into the UserPram struct
	var params models.UserPram
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "HandlerUsersCreate",
		}).Error("Couldn't decode parameters")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error("Failed to hash password")
		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create a new user in the database
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		Username:     params.UserName,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"func":     "HandlerUsersCreate",
			"userName": params.UserName,
		}).Error("Couldn't create user")
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	// Generate access and refresh tokens
	accessToken, refreshToken, err := cfg.Auth.LoginUser(params.UserName, params.Password)
	if err != nil {
		log.WithError(err).Warn("Invalid credentials")
		helper.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Prepare the response
	res := models.DatabaseUserToUser(user)
	res.AccessToken = accessToken
	res.RefreshToken = refreshToken

	// Respond with the newly created user and tokens
	w.WriteHeader(http.StatusCreated)
	helper.RespondWithJSON(w, http.StatusOK, res)
}

// HandlerGetUser retrieves and returns the user information.
func (cfg *ApiConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	helper.RespondWithJSON(w, http.StatusOK, models.DatabaseUserToUser(user))
}
