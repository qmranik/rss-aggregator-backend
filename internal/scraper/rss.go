package scraper

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

func ScrapeRSS(ctx context.Context, db *database.Queries, aiClient *ai.Client, sourceID uuid.UUID, sourceURL string) error {
	log.Printf("Fetching RSS feed from: %s", sourceURL)

	// Fetch and parse the feed data
	feedData, err := fetchFeed(sourceURL)
	if err != nil {
		return err
	}

	// Insert each post from the feed into the database
	for _, item := range feedData.Channel.Item {
		publishedAt, err := helper.ParsePubDate(item.PubDate)
		if err != nil {
			log.WithFields(log.Fields{
				"pubDate": item.PubDate,
				"url":     sourceURL,
				"error":   err,
			}).Error("Couldn't parse Published Date")
			continue
		}

		// Generate Summary and Embedding
		var aiSummary, rawContent sql.NullString
		var embedding interface{}
		if aiClient != nil {
			summary, err := aiClient.SummarizeContent(ctx, item.Description)
			if err != nil {
				log.Warnf("Failed to generate AI summary for RSS item: %v", err)
			} else {
				aiSummary = sql.NullString{String: summary, Valid: true}
			}
			rawContent = sql.NullString{String: item.Description, Valid: true}

			emb, err := aiClient.GenerateEmbedding(ctx, item.Description)
			if err != nil {
				log.Warnf("Failed to generate AI embedding for RSS item: %v", err)
			} else if emb != nil {
				embedding = pgvector.NewVector(emb)
			}
		}

		// Insert post
		_, err = db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: true},
			Url:         item.Link,
			PublishedAt: publishedAt,
			SourceID:    sourceID,
			RawContent:  rawContent,
			AiSummary:   aiSummary,
			Embedding:   embedding,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue // Skip duplicates
			}
			log.WithFields(log.Fields{
				"sourceID": sourceID,
				"title":    item.Title,
				"error":    err,
			}).Error("Couldn't create post")
		}
	}

	log.Infof("RSS Feed %s collected, %v posts found", sourceURL, len(feedData.Channel.Item))
	return nil
}

func fetchFeed(feedURL string) (*models.RSSFeed, error) {
	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed models.RSSFeed
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return nil, err
	}

	return &rssFeed, nil
}
