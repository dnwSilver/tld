package projectsync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/storage"
)

var ErrFileNotFound = errors.New("file not found")

type Service struct {
	Cache        storage.CacheRepository
	Runs         storage.ProjectDependencyRepository
	SourceClient SourceClient
	Force        bool
}

type Result struct {
	Run        storage.ProjectDependencyRun
	Count      int
	UpToDate   bool
	Dependency []storage.ProjectDependency
}

func (s Service) Sync(ctx context.Context, source Source, project Project, progress ProgressFunc) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	if s.SourceClient == nil {
		return Result{}, errors.New("project sync source client is empty")
	}
	if project.ID == 0 {
		return Result{}, errors.New("project sync project is empty")
	}
	if project.ProviderID == "" {
		return Result{}, errors.New("project sync provider project id is empty")
	}

	strategy, err := ResolveStackStrategy(project.StackName)
	if err != nil {
		return Result{}, err
	}

	report(progress, "Resolving commit...")
	commit, err := s.SourceClient.ResolveHead(ctx, source, project)
	if err != nil {
		return Result{}, err
	}
	if commit.ShortSHA == "" {
		commit.ShortSHA = shortSHA(commit.SHA)
	}

	hasRun, err := s.Runs.HasSuccessfulRun(ctx, project.ID, commit.ShortSHA)
	if err != nil {
		return Result{}, err
	}
	if hasRun && !s.Force {
		report(progress, "already up to date")
		dependencies, err := s.Runs.ListByProject(ctx, project.ID)
		if err != nil {
			return Result{}, err
		}
		run, err := s.Runs.LatestRun(ctx, project.ID)
		if err != nil {
			return Result{}, err
		}

		return Result{Run: run, Count: len(dependencies), UpToDate: true, Dependency: dependencies}, nil
	}

	files := make([]File, 0, len(strategy.Files()))
	for _, path := range strategy.Files() {
		content, err := s.fetchCachedFile(ctx, source, project, commit, path, progress)
		if errors.Is(err, ErrFileNotFound) {
			report(progress, fmt.Sprintf("Skipping missing %s", path))
			continue
		}
		if err != nil {
			return Result{}, err
		}
		files = append(files, File{Path: path, Content: content})
	}

	report(progress, "Parsing dependencies...")
	dependencies := make([]storage.ProjectDependency, 0)
	for _, file := range files {
		parsed, err := strategy.Parse(file.Path, file.Content)
		if err != nil {
			return Result{}, err
		}
		for index := range parsed {
			parsed[index].Name = safetext.Plain(parsed[index].Name)
			parsed[index].Version = safetext.Plain(parsed[index].Version)
			parsed[index].DependencyType = safetext.Plain(parsed[index].DependencyType)
			parsed[index].SourceFile = safetext.Plain(parsed[index].SourceFile)
			parsed[index].Ecosystem = dependencyEcosystem(parsed[index])
		}
		dependencies = append(dependencies, parsed...)
	}
	dependencies = preferNodeFromNvmrc(dependencies)

	report(progress, "Saving dependencies...")
	now := time.Now().UTC()
	run, err := s.Runs.ReplaceForProjectRun(ctx, project.ID, storage.ProjectDependencyRun{
		ProjectID:      project.ID,
		CommitShortSHA: commit.ShortSHA,
		CommitSHA:      commit.SHA,
		Status:         storage.ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, dependencies)
	if err != nil {
		return Result{}, err
	}

	report(progress, fmt.Sprintf("Saved %d dependencies", len(dependencies)))
	return Result{Run: run, Count: len(dependencies), Dependency: dependencies}, nil
}

func dependencyEcosystem(dependency storage.ProjectDependency) string {
	switch dependency.DependencyType {
	case DependencyTypeRuntime, DependencyTypeDev, DependencyTypePeer, DependencyTypeOptional:
		return "npm"
	case DependencyTypeEngines, DependencyTypeNvmrc:
		return "node"
	case DependencyTypeGoModule:
		return "go"
	case DependencyTypeGradleLibrary, DependencyTypeGradlePlugin:
		return "maven"
	case DependencyTypeCocoaPods:
		return "cocoapods"
	case DependencyTypeBundler:
		return "rubygems"
	default:
		return ""
	}
}

func (s Service) fetchCachedFile(ctx context.Context, source Source, project Project, commit Commit, path string, progress ProgressFunc) ([]byte, error) {
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

// sourceCacheKey separates provider installations and invalidates ambiguous legacy keys.
func sourceCacheKey(source Source, projectID, revision, kind string) string {
	parts := []string{"v2", source.Type, fmt.Sprint(source.ID), source.URL, projectID, revision, kind}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "v2:" + hex.EncodeToString(sum[:])
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) <= 8 {
		return sha
	}

	return sha[:8]
}

func report(progress ProgressFunc, message string) {
	if progress != nil {
		progress(message)
	}
}

func preferNodeFromNvmrc(dependencies []storage.ProjectDependency) []storage.ProjectDependency {
	hasNvmrcNode := false
	for _, dependency := range dependencies {
		if strings.EqualFold(strings.TrimSpace(dependency.Name), "node") && dependency.SourceFile == ".nvmrc" {
			hasNvmrcNode = true
			break
		}
	}
	if !hasNvmrcNode {
		return dependencies
	}

	filtered := make([]storage.ProjectDependency, 0, len(dependencies))
	for _, dependency := range dependencies {
		if strings.EqualFold(strings.TrimSpace(dependency.Name), "node") && dependency.SourceFile != ".nvmrc" {
			continue
		}
		filtered = append(filtered, dependency)
	}

	return filtered
}
