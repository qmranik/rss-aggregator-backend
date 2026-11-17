package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
)

type Source struct {
	ID            uuid.UUID  `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `json:"name"`
	Url           string     `json:"url"`
	UserID        uuid.UUID  `json:"user_id"`
	SourceType    string     `json:"source_type"`
	FetchInterval string     `json:"fetch_interval"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
}

func DatabaseSourceToSource(source database.Source) Source {
	return Source{
		ID:            source.ID,
		CreatedAt:     source.CreatedAt,
		UpdatedAt:     source.UpdatedAt,
		Name:          source.Name,
		Url:           source.Url,
		UserID:        source.UserID,
		SourceType:    string(source.SourceType),
		FetchInterval: string(source.FetchInterval),
		LastFetchedAt: NullTimeToTimePtr(source.LastFetchedAt),
	}
}

func DatabaseSourcesToSources(sources []database.Source) []Source {
	result := make([]Source, len(sources))
	for i, source := range sources {
		result[i] = DatabaseSourceToSource(source)
	}
	return result
}

type SourceFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	SourceID  uuid.UUID `json:"source_id"`
}

func DatabaseSourceFollowToSourceFollow(sourceFollow database.SourceFollow) SourceFollow {
	return SourceFollow{
		ID:        sourceFollow.ID,
		CreatedAt: sourceFollow.CreatedAt,
		UpdatedAt: sourceFollow.UpdatedAt,
		UserID:    sourceFollow.UserID,
		SourceID:  sourceFollow.SourceID,
	}
}

func DatabaseSourceFollowsToSourceFollows(sourceFollows []database.SourceFollow) []SourceFollow {
	result := make([]SourceFollow, len(sourceFollows))
	for i, sourceFollow := range sourceFollows {
		result[i] = DatabaseSourceFollowToSourceFollow(sourceFollow)
	}
	return result
}
