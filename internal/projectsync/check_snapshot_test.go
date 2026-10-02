package projectsync

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

type failingProcessModeClient struct{ *fakeSourceClient }

func (failingProcessModeClient) ResourceGroupProcessMode(context.Context, Source, Project, string) (string, error) {
	return "", errors.New("resource group unavailable")
}

func TestFailedCheckRunDoesNotExposeMixedSnapshot(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "checks.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	source := Source{Type: SourceTypeGitLab}
	project := Project{ID: 1, ProviderID: "group/repo", Name: "repo"}
	service := CheckService{Cache: store.Cache(), SourceClient: failingProcessModeClient{&fakeSourceClient{}}}
	previous := make(ProjectCheckResults, len(ProjectChecks))
	for _, check := range ProjectChecks {
		previous[check.ID] = CheckStatePass
	}
	if err := service.cacheSnapshot(ctx, source, project, previous, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RunProject(ctx, source, project, nil); err == nil {
		t.Fatal("expected a check failure")
	}
	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range ProjectChecks {
		if loaded[check.ID] != CheckStatePass {
			t.Fatalf("%s changed during failed run: %s", check.ID, loaded[check.ID])
		}
	}
	batch, _, err := service.LoadProjectsWithVersions(ctx, []ProjectSourceRef{
		{Source: source, Project: Project{ProviderID: "missing"}},
		{Source: source, Project: project},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != 2 || batch[0]["master"] != CheckStateUnknown || batch[1]["master"] != CheckStatePass {
		t.Fatalf("batch checks = %#v", batch)
	}
}
