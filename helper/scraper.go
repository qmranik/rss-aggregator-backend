package helper

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	log "github.com/sirupsen/logrus"
)

// StartScraping initiates a periodic feed scraping process.
// It fetches feeds every `timeBetweenRequest` duration using `concurrency` number of goroutines.
func StartScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Infof("Collecting feeds every %s using %v goroutines...", timeBetweenRequest, concurrency)
	ticker := time.NewTicker(timeBetweenRequest)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't get next feeds to fetch")
			continue
		}
		log.Infof("Found %v feeds to fetch!", len(feeds))

		var wg sync.WaitGroup
		for _, feed := range feeds {
			wg.Add(1)
			go ScrapeFeed(db, &wg, feed)
		}
		wg.Wait()
	}
}

// ScrapeFeed scrapes a single feed and inserts its posts into the database.
// It marks the feed as fetched before processing and handles potential errors.
func ScrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	// Mark the feed as fetched
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"feedID":   feed.ID,
			"feedName": feed.Name,
			"error":    err,
		}).Error("Couldn't mark feed fetched")
		return
	}

	// Fetch and parse the feed data
	feedData, err := FetchFeed(feed.Url)
	if err != nil {
		log.WithFields(log.Fields{
			"feedID":   feed.ID,
			"feedName": feed.Name,
			"feedUrl":  feed.Url,
			"error":    err,
		}).Error("Couldn't collect feed")
		return
	}

	// Insert each post from the feed into the database
	for _, item := range feedData.Channel.Item {
		publishedAt, err := ParsePubDate(item.PubDate)
		if err != nil {
			log.WithFields(log.Fields{
				"pubDate":  item.PubDate,
				"feedID":   feed.ID,
				"feedName": feed.Name,
				"error":    err,
			}).Error("Couldn't parse Published Date")
			continue
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			FeedID:      feed.ID,
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: true},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				// Skip duplicate entries
				continue
			}
			log.WithFields(log.Fields{
				"feedID":   feed.ID,
				"feedName": feed.Name,
				"title":    item.Title,
				"error":    err,
			}).Error("Couldn't create post")
		}
	}

	log.Infof("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

// FetchFeed retrieves and parses an RSS feed from the specified URL.
// It returns the parsed feed or an error if fetching or parsing fails.
func FetchFeed(feedURL string) (*models.RSSFeed, error) {
	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(feedURL)
	if err != nil {
		log.WithFields(log.Fields{
			"feedURL": feedURL,
			"error":   err,
		}).Error("Failed to fetch feed")
		return nil, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"feedURL": feedURL,
			"error":   err,
		}).Error("Failed to read response body")
		return nil, err
	}

	var rssFeed models.RSSFeed
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		log.WithFields(log.Fields{
			"feedURL": feedURL,
			"error":   err,
		}).Error("Failed to unmarshal XML")
		return nil, err
	}

	return &rssFeed, nil
}
