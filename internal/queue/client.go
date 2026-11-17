package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeScrapeRSS      = "scrape:rss"
	TypeScrapeHTML     = "scrape:html"
	TypeScrapeYouTube  = "scrape:youtube"
	TypeScrapeDocument = "scrape:document"
)

type ScrapePayload struct {
	SourceID string `json:"source_id"`
	URL      string `json:"url"`
}

type Client struct {
	asynqClient *asynq.Client
}

func NewClient(redisAddr string) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Client{asynqClient: client}
}

func (c *Client) EnqueueScrape(taskType string, payload ScrapePayload, delay time.Duration) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("could not marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, b)
	
	// Enqueue with optional delay
	info, err := c.asynqClient.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return fmt.Errorf("could not enqueue task: %w", err)
	}

	log.Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

func (c *Client) Close() error {
	return c.asynqClient.Close()
}
