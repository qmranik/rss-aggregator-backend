package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/internal/scraper"
)

type Worker struct {
	server   *asynq.Server
	db       *database.Queries
	aiClient *ai.Client
}

func NewWorker(redisAddr string, db *database.Queries, aiClient *ai.Client, concurrency int) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)
	return &Worker{
		server:   srv,
		db:       db,
		aiClient: aiClient,
	}
}

func (w *Worker) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeScrapeRSS, w.handleScrapeRSS)
	mux.HandleFunc(TypeScrapeHTML, w.handleScrapeHTML)
	mux.HandleFunc(TypeScrapeYouTube, w.handleScrapeYouTube)
	mux.HandleFunc(TypeScrapeDocument, w.handleScrapeDocument)

	log.Println("Starting Asynq worker server...")
	return w.server.Start(mux)
}

func (w *Worker) Stop() {
	log.Println("Shutting down Asynq worker server...")
	w.server.Shutdown()
}

func (w *Worker) handleScrapeRSS(ctx context.Context, t *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Processing RSS scrape for source: %s, url: %s", payload.SourceID, payload.URL)
	sourceUUID, err := uuid.Parse(payload.SourceID)
	if err != nil {
		return err
	}
	
	// Implementation will call the actual scraper logic
	return scraper.ScrapeRSS(ctx, w.db, w.aiClient, sourceUUID, payload.URL)
}

func (w *Worker) handleScrapeHTML(ctx context.Context, t *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Processing HTML scrape for source: %s, url: %s", payload.SourceID, payload.URL)
	sourceUUID, err := uuid.Parse(payload.SourceID)
	if err != nil {
		return err
	}
	return scraper.ScrapeHTML(ctx, w.db, w.aiClient, sourceUUID, payload.URL)
}

func (w *Worker) handleScrapeYouTube(ctx context.Context, t *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Processing YouTube scrape for source: %s, channel: %s", payload.SourceID, payload.URL)
	sourceUUID, err := uuid.Parse(payload.SourceID)
	if err != nil {
		return err
	}
	return scraper.ScrapeYouTube(ctx, w.db, w.aiClient, sourceUUID, payload.URL) // URL is used as channel ID here
}

func (w *Worker) handleScrapeDocument(ctx context.Context, t *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Processing Document scrape for source: %s, url: %s", payload.SourceID, payload.URL)
	return nil
}
