package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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
	{ID: checkProtectedBranches, Title: "branches"},
	{ID: checkProtectedTags, Title: "tags"},
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

const checkSnapshotVersion = 1

type checkSnapshot struct {
	Version   int                  `json:"version"`
	CheckedAt time.Time            `json:"checked_at"`
	Results   ProjectCheckResults  `json:"results"`
	Versions  ProjectCheckVersions `json:"versions"`
}

func (s CheckService) RunProject(ctx context.Context, source Source, project Project, progress ProgressFunc) (ProjectCheckResults, error) {
	if s.SourceClient == nil {
		return nil, errors.New("check source client is empty")
	}
	if project.ProviderID == "" {
		return nil, errors.New("check provider project id is empty")
	}

	var ciContent []byte
	var ciFound bool
	var err error
	gitlabSource := strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab)
	if gitlabSource {
		ciContent, ciFound, err = s.loadGitlabCI(ctx, source, project)
		if err != nil {
			return nil, err
		}
	}
	var protectedBranches []ProtectedBranch
	var protectedFound bool
	var ciSettings CISettings
	var ciSettingsFound bool
	if gitlabSource {
		protectedBranches, protectedFound, err = s.loadProtectedBranches(ctx, source, project)
		if err != nil {
			return nil, err
		}
		ciSettings, ciSettingsFound, err = s.loadCISettings(ctx, source, project)
		if err != nil {
			return nil, err
		}
	}
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
	}
	if err := s.cacheSnapshot(ctx, source, project, results, versions); err != nil {
		return nil, err
	}

	return results, nil
}

func (s CheckService) LoadProject(ctx context.Context, source Source, project Project) (ProjectCheckResults, error) {
	results, _, err := s.LoadProjectWithVersions(ctx, source, project)
	return results, err
}

func (s CheckService) LoadProjectWithVersions(ctx context.Context, source Source, project Project) (ProjectCheckResults, ProjectCheckVersions, error) {
	entry, err := s.Cache.Get(ctx, CacheNamespaceProjectChecks, checkSnapshotKey(source, project.ProviderID))
	if errors.Is(err, storage.ErrCacheMiss) {
		return unknownCheckResults(), make(ProjectCheckVersions), nil
	}
	if err != nil {
		return nil, nil, err
	}
	return decodeCheckSnapshot(entry.Value)
}

func (s CheckService) LoadProjectsWithVersions(ctx context.Context, refs []ProjectSourceRef) ([]ProjectCheckResults, []ProjectCheckVersions, error) {
	results := make([]ProjectCheckResults, len(refs))
	versions := make([]ProjectCheckVersions, len(refs))
	keys := make([]string, len(refs))
	for i, ref := range refs {
		keys[i] = checkSnapshotKey(ref.Source, ref.Project.ProviderID)
	}
	entries, err := s.Cache.GetMany(ctx, CacheNamespaceProjectChecks, keys)
	if err != nil {
		return nil, nil, err
	}
	for i, key := range keys {
		entry, ok := entries[key]
		if !ok {
			results[i] = unknownCheckResults()
			versions[i] = make(ProjectCheckVersions)
			continue
		}
		results[i], versions[i], err = decodeCheckSnapshot(entry.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("decode checks for project %q: %w", refs[i].Project.ProviderID, err)
		}
	}
	return results, versions, nil
}

func unknownCheckResults() ProjectCheckResults {
	results := make(ProjectCheckResults, len(ProjectChecks))
	for _, check := range ProjectChecks {
		results[check.ID] = CheckStateUnknown
	}
	return results
}

func decodeCheckSnapshot(value []byte) (ProjectCheckResults, ProjectCheckVersions, error) {
	results := make(ProjectCheckResults, len(ProjectChecks))
	versions := make(ProjectCheckVersions, len(ProjectChecks))
	var snapshot checkSnapshot
	if err := json.Unmarshal(value, &snapshot); err != nil {
		return nil, nil, fmt.Errorf("decode project checks snapshot: %w", err)
	}
	if snapshot.Version != checkSnapshotVersion {
		return nil, nil, fmt.Errorf("unsupported project checks snapshot version %d", snapshot.Version)
	}
	for _, check := range ProjectChecks {
		state, ok := snapshot.Results[check.ID]
		if !ok {
			return nil, nil, fmt.Errorf("project checks snapshot is missing %s", check.ID)
		}
		switch state {
		case CheckStatePass, CheckStateWarning, CheckStateFail, CheckStateNotApplicable, CheckStateUnknown:
			results[check.ID] = state
		default:
			return nil, nil, fmt.Errorf("project checks snapshot has invalid state %q for %s", state, check.ID)
		}
		versions[check.ID] = snapshot.Versions[check.ID]
	}

	return results, versions, nil
}

func (s CheckService) runCheck(ctx context.Context, source Source, project Project, checkID string, ciContent []byte, ciFound bool, protectedBranches []ProtectedBranch, protectedFound bool, ciSettings CISettings, ciSettingsFound bool, nextConfigFound bool, hasProductionNext bool) (CheckState, error) {
	if gitlabOnlyCheck(checkID) && !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
		return CheckStateNotApplicable, nil
	}
	if checkID == checkProtectedBranches || checkID == checkProtectedTags {
		return s.runProtectionCheck(ctx, source, project, checkID)
	}
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

func gitlabOnlyCheck(checkID string) bool {
	switch checkID {
	case checkProcessMode, checkSeparatedCaches, checkProtectedBranches, checkProtectedTags,
		checkNightly, "ci/cd", "ntfy", "dtrack", "cremr", "protect":
		return true
	default:
		return false
	}
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

func (s CheckService) loadGitlabCI(ctx context.Context, source Source, project Project) ([]byte, bool, error) {
	commit, err := s.SourceClient.ResolveHead(ctx, source, project)
	if err != nil {
		return nil, false, fmt.Errorf("resolve head for CI checks: %w", err)
	}
	content, err := s.SourceClient.FetchFile(ctx, source, project, commit.SHA, gitlabCIFile)
	if errors.Is(err, ErrFileNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("fetch %s: %w", gitlabCIFile, err)
	}
	return content, true, nil
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

func (s CheckService) loadProtectedBranches(ctx context.Context, source Source, project Project) ([]ProtectedBranch, bool, error) {
	branches, err := s.SourceClient.ProtectedBranches(ctx, source, project)
	if err != nil {
		return nil, false, fmt.Errorf("load protected branches: %w", err)
	}
	return branches, true, nil
}

func (s CheckService) loadCISettings(ctx context.Context, source Source, project Project) (CISettings, bool, error) {
	client, ok := s.SourceClient.(CISettingsSourceClient)
	if !ok {
		return CISettings{}, false, errors.New("CI settings source client is not supported")
	}
	settings, err := client.CISettings(ctx, source, project)
	if err != nil {
		return CISettings{}, false, fmt.Errorf("load CI settings: %w", err)
	}
	return settings, true, nil
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

func (s CheckService) cacheSnapshot(ctx context.Context, source Source, project Project, results ProjectCheckResults, versions ProjectCheckVersions) error {
	snapshot := checkSnapshot{Version: checkSnapshotVersion, CheckedAt: time.Now().UTC(), Results: results, Versions: versions}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode project checks snapshot: %w", err)
	}
	return s.Cache.Set(ctx, CacheNamespaceProjectChecks, checkSnapshotKey(source, project.ProviderID), payload, "application/json", resultCacheTTL)
}

func checkSnapshotKey(source Source, projectID string) string {
	return sourceCacheKey(source, projectID, "checks", "snapshot-v1")
}
