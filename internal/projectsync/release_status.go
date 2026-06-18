package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const CacheNamespaceProjectReleaseStatus = "project-release-status"

type ReleaseStatusIndicator string

const (
	ReleaseStatusUnknown      ReleaseStatusIndicator = "unknown"
	ReleaseStatusUntaggedMain ReleaseStatusIndicator = "untagged-main"
	ReleaseStatusDevAhead     ReleaseStatusIndicator = "dev-ahead"
	ReleaseStatusDevBehind    ReleaseStatusIndicator = "dev-behind"
)

type ReleaseStatusService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
}

func (s ReleaseStatusService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) ([]ReleaseStatusIndicator, error) {
	if s.SourceClient == nil {
		return nil, errors.New("release status source client is empty")
	}
	if project.ProviderID == "" {
		return nil, errors.New("release status provider project id is empty")
	}

	report(progress, fmt.Sprintf("Checking release status %s...", project.Name))

	tags, err := s.SourceClient.Tags(ctx, source, project)
	if err != nil {
		return nil, err
	}

	statuses := s.computeStatuses(ctx, source, project, tags)

	if err := s.cacheStatuses(ctx, source, project, statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

func (s ReleaseStatusService) LoadProject(ctx context.Context, source Source, project Project) ([]ReleaseStatusIndicator, error) {
	key := releaseStatusCacheKey(source.Type, project.ProviderID)
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectReleaseStatus, key)
	if err != nil {
		if errors.Is(err, storage.ErrCacheMiss) {
			return []ReleaseStatusIndicator{ReleaseStatusUnknown}, nil
		}
		return nil, err
	}

	var raw []string
	if err := json.Unmarshal(entry.Value, &raw); err != nil {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}, nil
	}

	result := make([]ReleaseStatusIndicator, 0, len(raw))
	for _, v := range raw {
		result = append(result, ReleaseStatusIndicator(v))
	}

	return result, nil
}

func (s ReleaseStatusService) computeStatuses(ctx context.Context, source Source, project Project, tags []Tag) []ReleaseStatusIndicator {
	versionFile := versionFileForStack(project.StackName)
	if versionFile == "" {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}
	}

	mainBranch := resolveMainBranch(ctx, s.SourceClient, source, project)
	if mainBranch == "" {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}
	}

	mainContent, err := s.SourceClient.FetchFile(ctx, source, project, mainBranch, versionFile)
	if err != nil {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}
	}

	mainVersion := readVersionFromContent(project.StackName, mainContent)
	if mainVersion == "" {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}
	}

	tagSet := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tagSet[tag.Name] = true
	}

	statuses := make([]ReleaseStatusIndicator, 0, 2)

	if !tagSet["v"+mainVersion] && !tagSet[mainVersion] {
		statuses = append(statuses, ReleaseStatusUntaggedMain)
	}

	hasDev, _ := s.SourceClient.HasBranch(ctx, source, project, "dev")
	if hasDev {
		divergence, err := s.SourceClient.CompareBranches(ctx, source, project, mainBranch, "dev")
		if err == nil {
			if divergence.Ahead > 0 {
				statuses = append(statuses, ReleaseStatusDevAhead)
			}
			if divergence.Behind > 0 {
				statuses = append(statuses, ReleaseStatusDevBehind)
			}
		}
	}

	if len(statuses) == 0 {
		return []ReleaseStatusIndicator{ReleaseStatusUnknown}
	}

	return statuses
}

func (s ReleaseStatusService) cacheStatuses(ctx context.Context, source Source, project Project, statuses []ReleaseStatusIndicator) error {
	raw := make([]string, len(statuses))
	for i, st := range statuses {
		raw[i] = string(st)
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("encode release statuses: %w", err)
	}
	key := releaseStatusCacheKey(source.Type, project.ProviderID)
	return s.Cache.Set(ctx, CacheNamespaceProjectReleaseStatus, key, payload, "application/json", 0)
}

func releaseStatusCacheKey(sourceType, projectID string) string {
	return cacheKey(sourceType, projectID, "status", "release")
}

func resolveMainBranch(ctx context.Context, client SourceClient, source Source, project Project) string {
	hasMaster, _ := client.HasBranch(ctx, source, project, "master")
	if hasMaster {
		return "master"
	}
	hasMain, _ := client.HasBranch(ctx, source, project, "main")
	if hasMain {
		return "main"
	}
	return ""
}

func versionFileForStack(stackName string) string {
	switch strings.ToLower(strings.TrimSpace(stackName)) {
	case "javascript", "js", "node", "nodejs", "node.js":
		return "package.json"
	case "kotlin", "android":
		return "gradle.properties"
	default:
		return ""
	}
}

func readVersionFromContent(stackName string, content []byte) string {
	switch strings.ToLower(strings.TrimSpace(stackName)) {
	case "javascript", "js", "node", "nodejs", "node.js":
		return readJSVersion(content)
	case "kotlin", "android":
		return readKotlinVersion(content)
	default:
		return ""
	}
}

func readJSVersion(content []byte) string {
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return ""
	}
	return strings.TrimSpace(pkg.Version)
}

var kotlinVersionPattern = regexp.MustCompile(`(?m)^\s*(?:VERSION_NAME|versionName|version)\s*=\s*(\S+)\s*$`)

func readKotlinVersion(content []byte) string {
	matches := kotlinVersionPattern.FindSubmatch(content)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(string(matches[1]))
}
