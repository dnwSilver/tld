package projectsync

import (
	"context"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	CacheNamespaceProjectFiles = "project-files"
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

type File struct {
	Path    string
	Content []byte
}

type Dependency = storage.ProjectDependency

type ProgressFunc func(message string)

type SourceClient interface {
	ResolveHead(ctx context.Context, source Source, project Project) (Commit, error)
	FetchFile(ctx context.Context, source Source, project Project, commitSHA string, path string) ([]byte, error)
}
