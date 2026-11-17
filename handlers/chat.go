package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pgvector/pgvector-go"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/internal/verify"
	log "github.com/sirupsen/logrus"
)

func (cfg *ApiConfig) HandlerChat(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Query string `json:"query"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if params.Query == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Query cannot be empty")
		return
	}

	// 1. Generate Embedding for the query
	aiClient, err := ai.NewClient(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to initialize AI client")
		return
	}

	queryEmbedding, err := aiClient.GenerateEmbedding(r.Context(), params.Query)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to generate embedding for query")
		return
	}

	// 2. Search database for relevant posts
	posts, err := cfg.DB.SearchPostsByVector(r.Context(), database.SearchPostsByVectorParams{
		UserID:    user.ID,
		Embedding: pgvector.NewVector(queryEmbedding),
		Limit:     5, // Top 5 most relevant
	})
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to search posts")
		return
	}

	// 3. Build labelled context. Each block carries the post id the model must cite
	// with; posts is already user-scoped by SearchPostsByVector, so this same set
	// doubles as the authorization boundary for the verifier below.
	var contextText string
	sources := make([]verify.Source, 0, len(posts))
	for _, post := range posts {
		body := ""
		if post.AiSummary.Valid {
			body = post.AiSummary.String
		} else if post.Description.Valid {
			body = post.Description.String
		}
		if body == "" {
			continue
		}
		contextText += "[post:" + post.ID.String() + "] " + body + "\n"
		sources = append(sources, verify.Source{
			ID:    post.ID.String(),
			Title: post.Title,
			URL:   post.Url,
		})
	}

	if len(sources) == 0 {
		helper.RespondWithJSON(w, http.StatusOK, map[string]any{
			"answer":   verify.SafeFallback,
			"verified": false,
		})
		return
	}

	answer, err := aiClient.AnswerQueryWithContext(r.Context(), params.Query, contextText)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to generate answer")
		return
	}

	// 4. Deterministic citation gate. The model supplied pointers; the database
	// supplies the words. An unresolvable pointer fails the entire answer rather
	// than returning a partially-grounded response.
	result, err := verify.NewGate(sources).Verify(answer)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":  user.ID,
			"rejected": result.Rejected,
		}).Warn("chat answer failed citation verification")
	}

	helper.RespondWithJSON(w, http.StatusOK, map[string]any{
		"answer":   result.Answer,
		"verified": result.Verified,
		"cited":    result.Cited,
	})
}
