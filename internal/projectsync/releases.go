package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dnwSilver/tld/internal/storage"
)

type ReleaseKind string

const (
	ReleaseKindRelease ReleaseKind = "release"
	ReleaseKindHotfix  ReleaseKind = "hotfix"
)

type Release struct {
	CreatedAt time.Time   `json:"created_at"`
	Kind      ReleaseKind `json:"kind"`
}

var releaseTagPattern = regexp.MustCompile(`^(v|release/|hotfix/)[0-9]+\.[0-9]+\.[0-9]+$`)

func releaseKindForTag(name string) (ReleaseKind, bool) {
	if !releaseTagPattern.MatchString(name) {
		return "", false
	}
	if strings.HasPrefix(name, "hotfix/") {
		return ReleaseKindHotfix, true
	}
	return ReleaseKindRelease, true
}

type ReleaseService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
}

func (s ReleaseService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) ([]Release, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
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

	dates := make([]Release, 0, len(tags))
	for _, tag := range tags {
		kind, ok := releaseKindForTag(tag.Name)
		if !ok || tag.CreatedAt.IsZero() {
			continue
		}
		dates = append(dates, Release{CreatedAt: tag.CreatedAt.UTC(), Kind: kind})
	}

	if err := s.cacheDates(ctx, source, project, dates); err != nil {
		return nil, err
	}

	return dates, nil
}

func (s ReleaseService) LoadProject(ctx context.Context, source Source, project Project) ([]Release, error) {
	key := releaseSourceCacheKey(source, project.ProviderID)
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectReleases, key)
	if err != nil {
		if errors.Is(err, storage.ErrCacheMiss) {
			return []Release{}, nil
		}
		return nil, err
	}

	var dates []Release
	if err := json.Unmarshal(entry.Value, &dates); err != nil {
		// Older cache entries stored dates without tag names or kinds.
		var legacy []time.Time
		if legacyErr := json.Unmarshal(entry.Value, &legacy); legacyErr != nil {
			return nil, fmt.Errorf("decode cached releases: %w", err)
		}
		dates = make([]Release, 0, len(legacy))
		for _, date := range legacy {
			dates = append(dates, Release{CreatedAt: date, Kind: ReleaseKindRelease})
		}
	}

	return dates, nil
}

func (s ReleaseService) cacheDates(ctx context.Context, source Source, project Project, dates []Release) error {
	payload, err := json.Marshal(dates)
	if err != nil {
		return fmt.Errorf("encode releases: %w", err)
	}
	key := releaseSourceCacheKey(source, project.ProviderID)

	return s.Cache.Set(ctx, CacheNamespaceProjectReleases, key, payload, "application/json", resultCacheTTL)
}

func releaseSourceCacheKey(source Source, projectID string) string {
	return sourceCacheKey(source, projectID, "tags", "releases")
}
