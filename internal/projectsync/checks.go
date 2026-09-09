package projectsync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	gitlabCIFile    = ".gitlab-ci.yml"
	nextConfigFile  = "next.config.ts"
	packageJSONFile = "package.json"
)

const (
	ciTemplatesMarker = "spectrum-frontend/ci-templates"
	ntfyMarker        = "$CI_SERVER_FQDN/shared/ci-ntfy/"
	cremrMarker       = "$CI_SERVER_FQDN/shared/ci-mr/create-mr@"
)

const (
	checkNightly               = "nightly"
	nightlyScheduleDescription = "🌚 Nightly build"
	nightlyScheduleRef         = "master"
	nightlyScheduleOwner       = "group349_bot2"
)

const (
	accessLevelDeveloper  = 30
	accessLevelMaintainer = 40
)

const (
	checkProcessMode       = "processMode"
	checkSeparatedCaches   = "separatedCaches"
	processModeOldestFirst = "oldest_first"
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
	CheckStateUnknown       CheckState = "unknown"
	CheckStatePass          CheckState = "pass"
	CheckStateWarning       CheckState = "warning"
	CheckStateFail          CheckState = "fail"
	CheckStateNotApplicable CheckState = "not-applicable"
)

const checkNext = "next"

var ProjectChecks = []CheckDefinition{
	{ID: "master", Title: "master"},
	{ID: "dev", Title: "dev"},
	{ID: "default", Title: "default"},
	{ID: "protect", Title: "protect"},
	{ID: checkProcessMode, Title: "process"},
	{ID: checkSeparatedCaches, Title: "caches"},
	{ID: "ci/cd", Title: "ci/cd"},
	{ID: "ntfy", Title: "ntfy"},
	{ID: "dtrack", Title: "dtrack"},
	{ID: "cremr", Title: "cremr"},
	{ID: checkNightly, Title: "nightly"},
	{ID: checkNext, Title: "next"},
}

type CheckService struct {
	Cache        storage.CacheRepository
	SourceClient SourceClient
}

type ProjectCheckResults map[string]CheckState

type ProjectCheckVersions map[string]string

func (s CheckService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) (ProjectCheckResults, error) {
	if s.SourceClient == nil {
		return nil, errors.New("check source client is empty")
	}
	if project.ProviderID == "" {
		return nil, errors.New("check provider project id is empty")
	}

	ciContent, ciFound := s.loadGitlabCI(ctx, source, project)
	protectedBranches, protectedFound := s.loadProtectedBranches(ctx, source, project)
	ciSettings, ciSettingsFound := s.loadCISettings(ctx, source, project)
	packageJSONContent, packageJSONFound, nextConfigFound, err := s.loadNextProjectFiles(ctx, source, project)
	if err != nil {
		return nil, err
	}
	hasProductionNext := false
	if packageJSONFound {
		hasProductionNext, err = packageJSONHasProductionNext(packageJSONContent)
		if err != nil {
			return nil, err
		}
	}

	results := make(ProjectCheckResults, len(ProjectChecks))
	versions := ciComponentVersions(ciContent)
	for _, check := range ProjectChecks {
		report(progress, fmt.Sprintf("Checking %s %s...", project.Name, check.Title))
		state, err := s.runCheck(ctx, source, project, check.ID, ciContent, ciFound, protectedBranches, protectedFound, ciSettings, ciSettingsFound, nextConfigFound, hasProductionNext)
		if err != nil {
			return nil, err
		}
		results[check.ID] = state
		if err := s.cacheResult(ctx, source, project, check.ID, state, versions[check.ID]); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (s CheckService) LoadProject(ctx context.Context, source Source, project Project) (ProjectCheckResults, error) {
	results, _, err := s.LoadProjectWithVersions(ctx, source, project)
	return results, err
}

func (s CheckService) LoadProjectWithVersions(ctx context.Context, source Source, project Project) (ProjectCheckResults, ProjectCheckVersions, error) {
	results := make(ProjectCheckResults, len(ProjectChecks))
	versions := make(ProjectCheckVersions, len(ProjectChecks))
	for _, check := range ProjectChecks {
		key := checkCacheKey(source.Type, project.ProviderID, check.ID)
		entry, err := s.Cache.Get(ctx, CacheNamespaceProjectChecks, key)
		if err != nil {
			if errors.Is(err, storage.ErrCacheMiss) {
				results[check.ID] = CheckStateUnknown
				continue
			}
			return nil, nil, err
		}
		stateValue, version := decodeCheckCacheValue(entry.Value)
		switch stateValue {
		case "true", string(CheckStatePass):
			results[check.ID] = CheckStatePass
		case string(CheckStateWarning):
			results[check.ID] = CheckStateWarning
		case "false", string(CheckStateFail):
			results[check.ID] = CheckStateFail
		case string(CheckStateNotApplicable):
			results[check.ID] = CheckStateNotApplicable
		default:
			results[check.ID] = CheckStateUnknown
		}
		if version != "" {
			versions[check.ID] = version
		}
	}

	return results, versions, nil
}

func (s CheckService) runCheck(ctx context.Context, source Source, project Project, checkID string, ciContent []byte, ciFound bool, protectedBranches []ProtectedBranch, protectedFound bool, ciSettings CISettings, ciSettingsFound bool, nextConfigFound bool, hasProductionNext bool) (CheckState, error) {
	if checkID == checkProcessMode {
		return s.runProcessModeCheck(ctx, source, project)
	}
	if checkID == checkSeparatedCaches {
		if !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
			return CheckStateNotApplicable, nil
		}
		if !ciSettingsFound {
			return CheckStateFail, nil
		}
		// Feature branches must share one runner cache with dev/master,
		// so separated caches being disabled is the healthy state.
		return checkState(!ciSettings.SeparatedCaches), nil
	}

	if checkID == checkNightly {
		if !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
			return CheckStateNotApplicable, nil
		}
		client, ok := s.SourceClient.(PipelineScheduleSourceClient)
		if !ok {
			return CheckStateFail, errors.New("pipeline schedules source client is not supported")
		}
		schedules, err := client.PipelineSchedules(ctx, source, project)
		if err != nil {
			return CheckStateFail, err
		}
		if len(schedules) == 0 {
			return CheckStateFail, nil
		}
		if nightlyScheduleConfigured(schedules) {
			return CheckStatePass, nil
		}
		return CheckStateWarning, nil
	}

	if checkID == checkNext {
		if !hasProductionNext {
			return CheckStateNotApplicable, nil
		}
		return checkState(nextConfigFound), nil
	}

	var pass bool
	var err error
	switch checkID {
	case "master":
		pass, err = s.SourceClient.HasBranch(ctx, source, project, "master")
	case "dev":
		pass, err = s.SourceClient.HasBranch(ctx, source, project, "dev")
	case "default":
		var branch string
		branch, err = s.SourceClient.DefaultBranch(ctx, source, project)
		if err != nil {
			return CheckStateFail, err
		}
		pass = strings.EqualFold(branch, "dev")
	case "ci/cd":
		pass = ciFound && !strings.Contains(string(ciContent), ciTemplatesMarker)
	case "ntfy":
		pass = ciFound && strings.Contains(string(ciContent), ntfyMarker)
	case "dtrack":
		_, pass = ciComponentVersions(ciContent)[checkDtrack]
		pass = ciFound && pass
	case "cremr":
		pass = ciFound && strings.Contains(string(ciContent), cremrMarker)
	case "protect":
		pass = protectedFound && protectedBranchesValid(protectedBranches)
	default:
		return CheckStateFail, fmt.Errorf("unsupported check %q", checkID)
	}
	if err != nil {
		return CheckStateFail, err
	}
	return checkState(pass), nil
}

func (s CheckService) runProcessModeCheck(ctx context.Context, source Source, project Project) (CheckState, error) {
	if !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
		return CheckStateNotApplicable, nil
	}
	client, ok := s.SourceClient.(ResourceGroupSourceClient)
	if !ok {
		return CheckStateFail, errors.New("resource groups source client is not supported")
	}
	resourceGroup, err := releaseCandidateResourceGroup(ctx, client, source, project)
	if err != nil {
		return CheckStateFail, err
	}
	processMode, err := client.ResourceGroupProcessMode(ctx, source, project, resourceGroup)
	if errors.Is(err, ErrFileNotFound) {
		return CheckStateFail, nil
	}
	if err != nil {
		return CheckStateFail, err
	}

	return checkState(processMode == processModeOldestFirst), nil
}

func checkState(pass bool) CheckState {
	if pass {
		return CheckStatePass
	}
	return CheckStateFail
}

func nightlyScheduleConfigured(schedules []PipelineSchedule) bool {
	for _, schedule := range schedules {
		description := strings.TrimSpace(schedule.Description)
		ref := strings.TrimPrefix(strings.TrimSpace(schedule.Ref), "refs/heads/")
		owner := strings.TrimPrefix(strings.TrimSpace(schedule.OwnerUsername), "@")
		if description == nightlyScheduleDescription &&
			ref == nightlyScheduleRef &&
			strings.EqualFold(owner, nightlyScheduleOwner) {
			return true
		}
	}
	return false
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

func (s CheckService) loadNextProjectFiles(ctx context.Context, source Source, project Project) ([]byte, bool, bool, error) {
	commit, err := s.SourceClient.ResolveHead(ctx, source, project)
	if err != nil {
		return nil, false, false, fmt.Errorf("resolve head for Next.js checks: %w", err)
	}

	packageJSONContent, err := s.SourceClient.FetchFile(ctx, source, project, commit.SHA, packageJSONFile)
	packageJSONFound := true
	if errors.Is(err, ErrFileNotFound) {
		packageJSONFound = false
	} else if err != nil {
		return nil, false, false, fmt.Errorf("fetch %s: %w", packageJSONFile, err)
	}

	_, err = s.SourceClient.FetchFile(ctx, source, project, commit.SHA, nextConfigFile)
	nextConfigFound := true
	if errors.Is(err, ErrFileNotFound) {
		nextConfigFound = false
	} else if err != nil {
		return nil, false, false, fmt.Errorf("fetch %s: %w", nextConfigFile, err)
	}

	return packageJSONContent, packageJSONFound, nextConfigFound, nil
}

func packageJSONHasProductionNext(content []byte) (bool, error) {
	dependencies, err := parsePackageJSON(content)
	if err != nil {
		return false, err
	}
	for _, dependency := range dependencies {
		if dependency.Name == "next" && dependency.DependencyType == DependencyTypeRuntime {
			return true, nil
		}
	}
	return false, nil
}

func (s CheckService) loadProtectedBranches(ctx context.Context, source Source, project Project) ([]ProtectedBranch, bool) {
	branches, err := s.SourceClient.ProtectedBranches(ctx, source, project)
	if err != nil {
		return nil, false
	}
	return branches, true
}

func (s CheckService) loadCISettings(ctx context.Context, source Source, project Project) (CISettings, bool) {
	client, ok := s.SourceClient.(CISettingsSourceClient)
	if !ok {
		return CISettings{}, false
	}
	settings, err := client.CISettings(ctx, source, project)
	if err != nil {
		return CISettings{}, false
	}
	return settings, true
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

func (s CheckService) cacheResult(ctx context.Context, source Source, project Project, checkID string, state CheckState, version string) error {
	value := "false"
	switch state {
	case CheckStatePass:
		value = "true"
	case CheckStateWarning:
		value = string(CheckStateWarning)
	case CheckStateNotApplicable:
		value = string(CheckStateNotApplicable)
	}
	version = strings.TrimSpace(version)
	if version != "" {
		value += "\t" + version
	}
	key := checkCacheKey(source.Type, project.ProviderID, checkID)
	return s.Cache.Set(ctx, CacheNamespaceProjectChecks, key, []byte(value), "text/plain", 0)
}

func decodeCheckCacheValue(value []byte) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(string(value)), "\t", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.TrimSpace(parts[1])
}

func checkCacheKey(sourceType string, projectID string, checkID string) string {
	parts := []string{
		strings.TrimSpace(sourceType),
		strings.TrimSpace(projectID),
		strings.TrimSpace(checkID),
	}
	return strings.Join(parts, ":")
}
