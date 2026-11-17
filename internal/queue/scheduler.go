package queue

import (
	"context"
	"time"

	"github.com/qmranik/rss-aggregator-backend/internal/database"
	log "github.com/sirupsen/logrus"
)

// StartScheduling polls the database for sources that need to be fetched and enqueues tasks.
func StartScheduling(db *database.Queries, qClient *Client, interval time.Duration) {
	log.Infof("Scheduling feeds every %s...", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		sources, err := db.GetNextSourcesToFetch(context.Background(), 10)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't get next sources to fetch")
			continue
		}
		
		if len(sources) > 0 {
			log.Infof("Found %v sources to schedule!", len(sources))
		}

		for _, source := range sources {
			// Mark the source as fetched so we don't enqueue it again immediately
			_, err := db.MarkSourceFetched(context.Background(), source.ID)
			if err != nil {
				log.Errorf("Couldn't mark source fetched: %v", err)
				continue
			}

			payload := ScrapePayload{
				SourceID: source.ID.String(),
				URL:      source.Url,
			}

			var taskType string
			switch source.SourceType {
			case database.SourceTypeEnumRss:
				taskType = TypeScrapeRSS
			case database.SourceTypeEnumHtml:
				taskType = TypeScrapeHTML
			case database.SourceTypeEnumYoutube:
				taskType = TypeScrapeYouTube
			case database.SourceTypeEnumDocument:
				taskType = TypeScrapeDocument
			default:
				log.Warnf("Unknown source type for source %s", source.ID)
				continue
			}

			err = qClient.EnqueueScrape(taskType, payload, 0)
			if err != nil {
				log.Errorf("Failed to enqueue task: %v", err)
			}
		}
	}
}
