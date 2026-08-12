package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

type VulnReport struct {
	Scanned bool
	Counts  VulnCounts
	Items   []Vulnerability
}

type VulnStrategy interface {
	Files() []string
	Scan(files []File) (VulnReport, error)
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
		return nil, fmt.Errorf("unsupported vulnerability scan for stack %q", stackName)
	}
}

type VulnScanService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
	Mode         VulnScanMode
}

func (s VulnScanService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) (VulnReport, error) {
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
	result, err := strategy.Scan(files)
	if err != nil {
		return VulnReport{}, err
	}
	result.Scanned = true

	if err := s.cacheReport(ctx, source, project, result); err != nil {
		return VulnReport{}, err
	}

	return result, nil
}

func (s VulnScanService) LoadProject(ctx context.Context, source Source, project Project) (VulnReport, error) {
	key := vulnCacheKey(source.Type, project.ProviderID, s.mode())
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectVulns, key)
	if err != nil {
		if errors.Is(err, storage.ErrCacheMiss) {
			return VulnReport{}, nil
		}
		return VulnReport{}, err
	}

	var report VulnReport
	if err := json.Unmarshal(entry.Value, &report); err != nil {
		return VulnReport{}, fmt.Errorf("decode cached vulnerabilities: %w", err)
	}

	return report, nil
}

func (s VulnScanService) fetchCachedFile(ctx context.Context, source Source, project Project, commit Commit, path string, progress ProgressFunc) ([]byte, error) {
	key := cacheKey(source.Type, project.ProviderID, commit.ShortSHA, path)
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
	if err := s.Cache.Set(ctx, CacheNamespaceProjectFiles, key, content, "application/octet-stream", 0); err != nil {
		return nil, err
	}

	return content, nil
}

func (s VulnScanService) cacheReport(ctx context.Context, source Source, project Project, report VulnReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("encode vulnerabilities: %w", err)
	}

	return s.Cache.Set(ctx, CacheNamespaceProjectVulns, vulnCacheKey(source.Type, project.ProviderID, s.mode()), payload, "application/json", 0)
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

func vulnCacheKey(sourceType string, projectID string, mode VulnScanMode) string {
	return cacheKey(sourceType, projectID, "vulns", string(normalizeVulnScanMode(mode))+"/report")
}
