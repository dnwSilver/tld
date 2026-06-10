package projectsync

import (
	"context"
	"time"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	CacheNamespaceProjectFiles    = "project-files"
	CacheNamespaceProjectChecks   = "project-checks"
	CacheNamespaceProjectReleases = "project-releases"
	CacheNamespaceProjectVulns    = "project-vulnerabilities"
)

type Project struct {
	ID         int64
	ProviderID string
	Name       string
	StackName  string
}

type Source struct {
	ID       int64
	Type     string
	URL      string
	PATToken string
}

type Commit struct {
	SHA      string
	ShortSHA string
}

type Tag struct {
	Name      string
	CreatedAt time.Time
}

type ProtectedBranch struct {
	Name              string
	PushAccessLevels  []int
	MergeAccessLevels []int
	AllowForcePush    bool
}

type File struct {
	Path    string
	Content []byte
}

type Dependency = storage.ProjectDependency

type ProgressFunc func(message string)

type SourceClient interface {
	ResolveHead(ctx context.Context, source Source, project Project) (Commit, error)
	FetchFile(ctx context.Context, source Source, project Project, commitSHA string, path string) ([]byte, error)
	HasBranch(ctx context.Context, source Source, project Project, branch string) (bool, error)
	DefaultBranch(ctx context.Context, source Source, project Project) (string, error)
	ProtectedBranches(ctx context.Context, source Source, project Project) ([]ProtectedBranch, error)
	Tags(ctx context.Context, source Source, project Project) ([]Tag, error)
}
