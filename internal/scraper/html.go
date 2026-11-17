package scraper

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	log "github.com/sirupsen/logrus"
)

func ScrapeHTML(ctx context.Context, db *database.Queries, aiClient *ai.Client, sourceID uuid.UUID, sourceURL string) error {
	log.Printf("Scraping HTML from: %s", sourceURL)

	res, err := http.Get(sourceURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	title := doc.Find("title").Text()
	
	// Just getting the main content of body
	// Real implementation would be more robust, extracting main article etc.
	content := doc.Find("body").Text() 
	
	// Truncate if too long for description
	desc := content
	if len(desc) > 500 {
		desc = desc[:500] + "..."
	}

	// Generate Summary and Embedding
	var aiSummary, rawContent sql.NullString
	var embedding interface{}
	if aiClient != nil {
		summary, err := aiClient.SummarizeContent(ctx, content)
		if err != nil {
			log.Warnf("Failed to generate AI summary for HTML item: %v", err)
		} else {
			aiSummary = sql.NullString{String: summary, Valid: true}
		}
		rawContent = sql.NullString{String: content, Valid: true}

		emb, err := aiClient.GenerateEmbedding(ctx, content)
		if err != nil {
			log.Warnf("Failed to generate AI embedding for HTML item: %v", err)
		} else if emb != nil {
			embedding = pgvector.NewVector(emb)
		}
	}

	_, err = db.CreatePost(ctx, database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Title:       title,
		Description: sql.NullString{String: desc, Valid: true},
		Url:         sourceURL,
		PublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		SourceID:    sourceID,
		RawContent:  rawContent,
		AiSummary:   aiSummary,
		Embedding:   embedding,
	})
	if err != nil {
		// Ignore duplicates
		return nil 
	}

	return nil
}
