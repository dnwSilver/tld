package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnwSilver/tld/internal/storage"
)

type ReleaseService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
}

func (s ReleaseService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) ([]time.Time, error) {
	if s.SourceClient == nil {
		return nil, errors.New("release source client is empty")
	}
	if project.ProviderID == "" {
		return nil, errors.New("release provider project id is empty")
	}

	report(progress, fmt.Sprintf("Loading releases %s...", project.Name))
	tags, err := s.SourceClient.Tags(ctx, source, project)
	if err != nil {
		return nil, err
	}

	dates := make([]time.Time, 0, len(tags))
	for _, tag := range tags {
		if tag.CreatedAt.IsZero() {
			continue
		}
		dates = append(dates, tag.CreatedAt.UTC())
	}

	if err := s.cacheDates(ctx, source, project, dates); err != nil {
		return nil, err
	}

	return dates, nil
}

func (s ReleaseService) LoadProject(ctx context.Context, source Source, project Project) ([]time.Time, error) {
	key := releaseCacheKey(source.Type, project.ProviderID)
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectReleases, key)
	if err != nil {
		if errors.Is(err, storage.ErrCacheMiss) {
			return []time.Time{}, nil
		}
		return nil, err
	}

	var dates []time.Time
	if err := json.Unmarshal(entry.Value, &dates); err != nil {
		return nil, fmt.Errorf("decode cached releases: %w", err)
	}

	return dates, nil
}

func (s ReleaseService) cacheDates(ctx context.Context, source Source, project Project, dates []time.Time) error {
	payload, err := json.Marshal(dates)
	if err != nil {
		return fmt.Errorf("encode releases: %w", err)
	}
	key := releaseCacheKey(source.Type, project.ProviderID)

	return s.Cache.Set(ctx, CacheNamespaceProjectReleases, key, payload, "application/json", 0)
}

func releaseCacheKey(sourceType string, projectID string) string {
	return cacheKey(sourceType, projectID, "tags", "releases")
}
