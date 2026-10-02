package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/storage"
)

type VulnSeverity string

const (
	VulnSeverityCritical VulnSeverity = "critical"
	VulnSeverityHigh     VulnSeverity = "high"
	VulnSeverityMedium   VulnSeverity = "medium"
	VulnSeverityLow      VulnSeverity = "low"
	VulnSeverityNone     VulnSeverity = "none"
)

type VulnScanMode string

const (
	VulnScanModeProd VulnScanMode = "prod"
	VulnScanModeDev  VulnScanMode = "dev"
)

type VulnCounts struct {
	Critical int
	High     int
	Medium   int
	Low      int
	None     int
}

type Vulnerability struct {
	Package     string
	Severity    VulnSeverity
	Title       string
	Description string
	Range       string
}

type VulnScanOutcome string

const (
	VulnScanComplete    VulnScanOutcome = "complete"
	VulnScanFailed      VulnScanOutcome = "failed"
	VulnScanCanceled    VulnScanOutcome = "canceled"
	VulnScanUnsupported VulnScanOutcome = "unsupported"
)

var ErrUnsupportedVulnStrategy = errors.New("unsupported vulnerability scan")

type VulnScanAttempt struct {
	Outcome VulnScanOutcome
	At      time.Time
}

type ScannerCoverage struct {
	Scanner  string
	Method   string
	Packages int
	Targets  int
}

type VulnReport struct {
	Scanned     bool
	Outcome     VulnScanOutcome
	Revision    string
	ScannedAt   time.Time
	Coverage    []ScannerCoverage
	Counts      VulnCounts
	Items       []Vulnerability
	LastAttempt VulnScanAttempt `json:"-"`
}

type VulnStrategy interface {
	Files() []string
	Scan(files []File) (VulnReport, error)
}

type contextVulnStrategy interface {
	ScanContext(ctx context.Context, files []File) (VulnReport, error)
}

func ResolveVulnStrategy(stackName string, mode VulnScanMode) (VulnStrategy, error) {
	normalized := strings.ToLower(strings.TrimSpace(stackName))
	mode = normalizeVulnScanMode(mode)
	switch normalized {
	case "javascript", "js", "node", "nodejs", "node.js":
		return JavaScriptVulnStrategy{Mode: mode}, nil
	case "kotlin", "android":
		return AndroidVulnStrategy{}, nil
	case "swift", "ios":
		return IOSVulnStrategy{}, nil
	default:
		return nil, fmt.Errorf("%w for stack %q", ErrUnsupportedVulnStrategy, stackName)
	}
}

type VulnScanService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
	Mode         VulnScanMode
}

type VulnProjectRef = ProjectSourceRef

func (s VulnScanService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) (VulnReport, error) {
	result, err := s.runProjectOnce(ctx, source, project, progress)
	if project.ProviderID == "" {
		return result, err
	}
	outcome := VulnScanComplete
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		outcome = VulnScanCanceled
	case errors.Is(err, ErrUnsupportedVulnStrategy):
		outcome = VulnScanUnsupported
	case err != nil:
		outcome = VulnScanFailed
	}
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	attemptErr := s.cacheAttempt(writeCtx, source, project, VulnScanAttempt{Outcome: outcome, At: time.Now().UTC()})
	if err != nil {
		return result, err
	}
	if attemptErr != nil {
		return VulnReport{}, attemptErr
	}
	return result, nil
}

func (s VulnScanService) runProjectOnce(ctx context.Context, source Source, project Project, progress ProgressFunc) (VulnReport, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	if s.SourceClient == nil {
		return VulnReport{}, errors.New("vulnerability source client is empty")
	}
	if project.ProviderID == "" {
		return VulnReport{}, errors.New("vulnerability provider project id is empty")
	}

	strategy, err := ResolveVulnStrategy(project.StackName, s.mode())
	if err != nil {
		return VulnReport{}, err
	}

	report(progress, "Resolving commit...")
	commit, err := s.SourceClient.ResolveHead(ctx, source, project)
	if err != nil {
		return VulnReport{}, err
	}
	if commit.ShortSHA == "" {
		commit.ShortSHA = shortSHA(commit.SHA)
	}

	files := make([]File, 0, len(strategy.Files()))
	for _, path := range strategy.Files() {
		content, err := s.fetchCachedFile(ctx, source, project, commit, path, progress)
		if errors.Is(err, ErrFileNotFound) {
			report(progress, fmt.Sprintf("Skipping missing %s", path))
			continue
		}
		if err != nil {
			return VulnReport{}, err
		}
		files = append(files, File{Path: path, Content: content})
	}

	report(progress, "Running vulnerability scan...")
	var result VulnReport
	if contextual, ok := strategy.(contextVulnStrategy); ok {
		result, err = contextual.ScanContext(ctx, files)
	} else {
		result, err = strategy.Scan(files)
	}
	if err != nil {
		return VulnReport{}, err
	}
	if !result.Scanned {
		return VulnReport{}, errors.New("vulnerability scanner returned no coverage")
	}
	result.Outcome = VulnScanComplete
	result.Revision = commit.SHA
	result.ScannedAt = time.Now().UTC()
	sanitizeVulnerabilityItems(result.Items)

	if err := s.cacheReport(ctx, source, project, result); err != nil {
		return VulnReport{}, err
	}

	return result, nil
}

func (s VulnScanService) LoadProject(ctx context.Context, source Source, project Project) (VulnReport, error) {
	reports, err := s.LoadProjects(ctx, []VulnProjectRef{{Source: source, Project: project}})
	if err != nil {
		return VulnReport{}, err
	}
	return reports[0], nil
}

// LoadProjects reads all visible project reports through one cache query.
// A missing report is represented by a zero-value VulnReport at its input index.
func (s VulnScanService) LoadProjects(ctx context.Context, projects []VulnProjectRef) ([]VulnReport, error) {
	reports := make([]VulnReport, len(projects))
	keys := make([]string, len(projects))
	for i, ref := range projects {
		keys[i] = vulnSourceCacheKey(ref.Source, ref.Project.ProviderID, s.mode())
	}
	entries, err := s.Cache.GetMany(ctx, CacheNamespaceProjectVulns, keys)
	if err != nil {
		return nil, err
	}
	attemptKeys := make([]string, len(projects))
	for i, ref := range projects {
		attemptKeys[i] = vulnAttemptCacheKey(ref.Source, ref.Project.ProviderID, s.mode())
	}
	attemptEntries, err := s.Cache.GetMany(ctx, CacheNamespaceProjectVulns, attemptKeys)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		entry, ok := entries[key]
		if ok {
			if err := json.Unmarshal(entry.Value, &reports[i]); err != nil {
				return nil, fmt.Errorf("decode cached vulnerabilities for project %q: %w", projects[i].Project.ProviderID, err)
			}
			sanitizeVulnerabilityItems(reports[i].Items)
		}
		if attemptEntry, ok := attemptEntries[attemptKeys[i]]; ok {
			if err := json.Unmarshal(attemptEntry.Value, &reports[i].LastAttempt); err != nil {
				return nil, fmt.Errorf("decode scan attempt for project %q: %w", projects[i].Project.ProviderID, err)
			}
		}
	}
	return reports, nil
}

func sanitizeVulnerabilityItems(items []Vulnerability) {
	for index := range items {
		items[index].Package = safetext.Plain(items[index].Package)
		items[index].Title = safetext.Plain(items[index].Title)
		items[index].Description = safetext.Plain(items[index].Description)
		items[index].Range = safetext.Plain(items[index].Range)
	}
}

func (s VulnScanService) fetchCachedFile(ctx context.Context, source Source, project Project, commit Commit, path string, progress ProgressFunc) ([]byte, error) {
	key := sourceCacheKey(source, project.ProviderID, commit.SHA, path)
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectFiles, key)
	if err == nil {
		report(progress, fmt.Sprintf("Cache hit %s", path))
		return entry.Value, nil
	}
	if !errors.Is(err, storage.ErrCacheMiss) {
		return nil, err
	}

	report(progress, fmt.Sprintf("Downloading %s...", path))
	content, err := s.SourceClient.FetchFile(ctx, source, project, commit.SHA, path)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.Set(ctx, CacheNamespaceProjectFiles, key, content, "application/octet-stream", fileCacheTTL); err != nil {
		return nil, err
	}

	return content, nil
}

func (s VulnScanService) cacheReport(ctx context.Context, source Source, project Project, report VulnReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("encode vulnerabilities: %w", err)
	}

	return s.Cache.Set(ctx, CacheNamespaceProjectVulns, vulnSourceCacheKey(source, project.ProviderID, s.mode()), payload, "application/json", resultCacheTTL)
}

func (s VulnScanService) cacheAttempt(ctx context.Context, source Source, project Project, attempt VulnScanAttempt) error {
	payload, err := json.Marshal(attempt)
	if err != nil {
		return fmt.Errorf("encode vulnerability scan attempt: %w", err)
	}
	return s.Cache.Set(ctx, CacheNamespaceProjectVulns, vulnAttemptCacheKey(source, project.ProviderID, s.mode()), payload, "application/json", resultCacheTTL)
}

func (s VulnScanService) mode() VulnScanMode {
	return normalizeVulnScanMode(s.Mode)
}

func normalizeVulnScanMode(mode VulnScanMode) VulnScanMode {
	if mode == VulnScanModeDev {
		return VulnScanModeDev
	}
	return VulnScanModeProd
}

func vulnSourceCacheKey(source Source, projectID string, mode VulnScanMode) string {
	return sourceCacheKey(source, projectID, "vulns", string(normalizeVulnScanMode(mode))+"/report")
}

func vulnAttemptCacheKey(source Source, projectID string, mode VulnScanMode) string {
	return sourceCacheKey(source, projectID, "vulns", string(normalizeVulnScanMode(mode))+"/attempt")
}
