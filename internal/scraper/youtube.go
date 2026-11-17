package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func ScrapeYouTube(ctx context.Context, db *database.Queries, aiClient *ai.Client, sourceID uuid.UUID, channelID string) error {
	log.Printf("Scraping YouTube for channel: %s", channelID)
	
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("YOUTUBE_API_KEY not set")
	}

	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}

	call := youtubeService.Search.List([]string{"snippet"}).
		ChannelId(channelID).
		Order("date").
		MaxResults(5).
		Type("video")

	response, err := call.Do()
	if err != nil {
		return err
	}

	for _, item := range response.Items {
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		videoURL := "https://www.youtube.com/watch?v=" + item.Id.VideoId

		// Generate Summary and Embedding
		var aiSummary, rawContent sql.NullString
		var embedding interface{}
		if aiClient != nil {
			summary, err := aiClient.SummarizeContent(ctx, item.Snippet.Description)
			if err != nil {
				log.Warnf("Failed to generate AI summary for YouTube item: %v", err)
			} else {
				aiSummary = sql.NullString{String: summary, Valid: true}
			}
			rawContent = sql.NullString{String: item.Snippet.Description, Valid: true}

			emb, err := aiClient.GenerateEmbedding(ctx, item.Snippet.Description)
			if err != nil {
				log.Warnf("Failed to generate AI embedding for YouTube item: %v", err)
			} else if emb != nil {
				embedding = pgvector.NewVector(emb)
			}
		}

		_, err = db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Snippet.Title,
			Description: sql.NullString{String: item.Snippet.Description, Valid: true},
			Url:         videoURL,
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: true},
			SourceID:    sourceID,
			RawContent:  rawContent,
			AiSummary:   aiSummary,
			Embedding:   embedding,
		})
		if err != nil {
			// Skip duplicates
			continue
		}
	}

	return nil
}
