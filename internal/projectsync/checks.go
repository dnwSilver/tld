package projectsync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const gitlabCIFile = ".gitlab-ci.yml"

const (
	ciTemplatesMarker = "spectrum-frontend/ci-templates"
	ntfyMarker        = "$CI_SERVER_FQDN/shared/ci-ntfy/"
	dtrackMarker      = "/dtrack@"
	cremrMarker       = "$CI_SERVER_FQDN/shared/ci-mr/create-mr@"
)

const (
	accessLevelDeveloper  = 30
	accessLevelMaintainer = 40
)

var protectedBranchRules = map[string]ProtectedBranch{
	"dev": {
		MergeAccessLevels: []int{accessLevelDeveloper},
		PushAccessLevels:  []int{accessLevelMaintainer},
		AllowForcePush:    false,
	},
	"master": {
		MergeAccessLevels: []int{accessLevelDeveloper},
		PushAccessLevels:  []int{accessLevelMaintainer},
		AllowForcePush:    false,
	},
}

type CheckDefinition struct {
	ID    string
	Title string
}

type CheckState string

const (
	CheckStateUnknown CheckState = "unknown"
	CheckStatePass    CheckState = "pass"
	CheckStateFail    CheckState = "fail"
)

var ProjectChecks = []CheckDefinition{
	{ID: "master", Title: "master"},
	{ID: "dev", Title: "dev"},
	{ID: "default", Title: "default"},
	{ID: "protect", Title: "protect"},
	{ID: "ci/cd", Title: "ci/cd"},
	{ID: "ntfy", Title: "ntfy"},
	{ID: "dtrack", Title: "dtrack"},
	{ID: "cremr", Title: "cremr"},
}

type CheckService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
}

type ProjectCheckResults map[string]CheckState

func (s CheckService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) (ProjectCheckResults, error) {
	if s.SourceClient == nil {
		return nil, errors.New("check source client is empty")
	}
	if project.ProviderID == "" {
		return nil, errors.New("check provider project id is empty")
	}

	ciContent, ciFound := s.loadGitlabCI(ctx, source, project)
	protectedBranches, protectedFound := s.loadProtectedBranches(ctx, source, project)

	results := make(ProjectCheckResults, len(ProjectChecks))
	for _, check := range ProjectChecks {
		report(progress, fmt.Sprintf("Checking %s %s...", project.Name, check.Title))
		pass, err := s.runCheck(ctx, source, project, check.ID, ciContent, ciFound, protectedBranches, protectedFound)
		if err != nil {
			return nil, err
		}
		state := CheckStateFail
		if pass {
			state = CheckStatePass
		}
		results[check.ID] = state
		if err := s.cacheResult(ctx, source, project, check.ID, pass); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (s CheckService) LoadProject(ctx context.Context, source Source, project Project) (ProjectCheckResults, error) {
	results := make(ProjectCheckResults, len(ProjectChecks))
	for _, check := range ProjectChecks {
		key := checkCacheKey(source.Type, project.ProviderID, check.ID)
		entry, err := s.Cache.Get(ctx, CacheNamespaceProjectChecks, key)
		if err != nil {
			if errors.Is(err, storage.ErrCacheMiss) {
				results[check.ID] = CheckStateUnknown
				continue
			}
			return nil, err
		}
		switch strings.TrimSpace(string(entry.Value)) {
		case "true":
			results[check.ID] = CheckStatePass
		case "false":
			results[check.ID] = CheckStateFail
		default:
			results[check.ID] = CheckStateUnknown
		}
	}

	return results, nil
}

func (s CheckService) runCheck(ctx context.Context, source Source, project Project, checkID string, ciContent []byte, ciFound bool, protectedBranches []ProtectedBranch, protectedFound bool) (bool, error) {
	switch checkID {
	case "master":
		return s.SourceClient.HasBranch(ctx, source, project, "master")
	case "dev":
		return s.SourceClient.HasBranch(ctx, source, project, "dev")
	case "default":
		branch, err := s.SourceClient.DefaultBranch(ctx, source, project)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(branch, "dev"), nil
	case "ci/cd":
		return ciFound && !strings.Contains(string(ciContent), ciTemplatesMarker), nil
	case "ntfy":
		return ciFound && strings.Contains(string(ciContent), ntfyMarker), nil
	case "dtrack":
		return ciFound && strings.Contains(string(ciContent), dtrackMarker), nil
	case "cremr":
		return ciFound && strings.Contains(string(ciContent), cremrMarker), nil
	case "protect":
		return protectedFound && protectedBranchesValid(protectedBranches), nil
	default:
		return false, fmt.Errorf("unsupported check %q", checkID)
	}
}

func (s CheckService) loadGitlabCI(ctx context.Context, source Source, project Project) ([]byte, bool) {
	commit, err := s.SourceClient.ResolveHead(ctx, source, project)
	if err != nil {
		return nil, false
	}
	content, err := s.SourceClient.FetchFile(ctx, source, project, commit.SHA, gitlabCIFile)
	if err != nil {
		return nil, false
	}
	return content, true
}

func (s CheckService) loadProtectedBranches(ctx context.Context, source Source, project Project) ([]ProtectedBranch, bool) {
	branches, err := s.SourceClient.ProtectedBranches(ctx, source, project)
	if err != nil {
		return nil, false
	}
	return branches, true
}

func protectedBranchesValid(branches []ProtectedBranch) bool {
	for name, rule := range protectedBranchRules {
		branch, ok := findProtectedBranch(branches, name)
		if !ok {
			return false
		}
		if branch.AllowForcePush != rule.AllowForcePush {
			return false
		}
		if !equalIntSets(branch.MergeAccessLevels, rule.MergeAccessLevels) {
			return false
		}
		if !equalIntSets(branch.PushAccessLevels, rule.PushAccessLevels) {
			return false
		}
	}
	return true
}

func findProtectedBranch(branches []ProtectedBranch, name string) (ProtectedBranch, bool) {
	for _, branch := range branches {
		if branch.Name == name {
			return branch, true
		}
	}
	return ProtectedBranch{}, false
}

func equalIntSets(actual []int, expected []int) bool {
	if len(actual) != len(expected) {
		return false
	}
	counts := make(map[int]int, len(expected))
	for _, value := range expected {
		counts[value]++
	}
	for _, value := range actual {
		counts[value]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func (s CheckService) cacheResult(ctx context.Context, source Source, project Project, checkID string, pass bool) error {
	value := "false"
	if pass {
		value = "true"
	}
	key := checkCacheKey(source.Type, project.ProviderID, checkID)
	return s.Cache.Set(ctx, CacheNamespaceProjectChecks, key, []byte(value), "text/plain", 0)
}

func checkCacheKey(sourceType string, projectID string, checkID string) string {
	parts := []string{
		strings.TrimSpace(sourceType),
		strings.TrimSpace(projectID),
		strings.TrimSpace(checkID),
	}
	return strings.Join(parts, ":")
}
