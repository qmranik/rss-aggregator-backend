package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
)

// Post represents a blog post or feed item.
type Post struct {
	ID          uuid.UUID  `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Title       string     `json:"title"`
	Url         string     `json:"url"`
	Description *string    `json:"description"`
	PublishedAt *time.Time `json:"published_at"`
	FeedID      uuid.UUID  `json:"feed_id"`
}

// DatabasePostToPost converts a database.Post to a Post model.
func DatabasePostToPost(post database.Post) Post {
	return Post{
		ID:          post.ID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Title:       post.Title,
		Url:         post.Url,
		Description: NullStringToStringPtr(post.Description),
		PublishedAt: NullTimeToTimePtr(post.PublishedAt),
		FeedID:      post.FeedID,
	}
}

// DatabasePostsToPosts converts a slice of database.Post to a slice of Post models.
func DatabasePostsToPosts(posts []database.Post) []Post {
	result := make([]Post, len(posts))
	for i, post := range posts {
		result[i] = DatabasePostToPost(post)
	}
	return result
}
