package projectsync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestPackageJSONHasProductionNext(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "production dependency", content: `{"dependencies":{"next":"^15.0.0"}}`, want: true},
		{name: "development dependency", content: `{"devDependencies":{"next":"^15.0.0"}}`, want: false},
		{name: "peer dependency", content: `{"peerDependencies":{"next":"^15.0.0"}}`, want: false},
		{name: "different package", content: `{"dependencies":{"nextjs":"^0.0.3"}}`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := packageJSONHasProductionNext([]byte(test.content))
			if err != nil {
				t.Fatalf("parse package.json: %v", err)
			}
			if got != test.want {
				t.Fatalf("has production next = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCheckServiceCachesNextConfigResults(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	client := &fakeSourceClient{files: map[string][]byte{
		packageJSONFile: []byte(`{"dependencies":{"next":"^15.0.0"}}`),
		nextConfigFile:  []byte(`export default {}`),
	}}
	service := CheckService{Cache: store.Cache(), SourceClient: client}
	source := Source{Type: storage.SourceTypeGitLab}
	project := Project{ID: 1, ProviderID: "group/repo", Name: "Repo"}

	results, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatalf("run project checks: %v", err)
	}
	if results[checkNext] != CheckStatePass {
		t.Errorf("result %s = %q, want %q", checkNext, results[checkNext], CheckStatePass)
	}
	if client.fetches[nextConfigFile] != 1 {
		t.Fatalf("next.config.ts fetches = %d, want 1", client.fetches[nextConfigFile])
	}
	if client.fetches[packageJSONFile] != 1 {
		t.Fatalf("package.json fetches = %d, want 1", client.fetches[packageJSONFile])
	}

	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if loaded[checkNext] != CheckStatePass {
		t.Fatalf("loaded next = %q, want %q", loaded[checkNext], CheckStatePass)
	}
}

func TestCheckServiceFailsNextCheckWhenConfigIsMissing(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	client := &fakeSourceClient{files: map[string][]byte{
		packageJSONFile: []byte(`{"dependencies":{"next":"^15.0.0"}}`),
	}}
	service := CheckService{Cache: store.Cache(), SourceClient: client}
	source := Source{Type: storage.SourceTypeGitLab}
	project := Project{ID: 1, ProviderID: "group/repo", Name: "Repo"}

	results, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatalf("run project checks: %v", err)
	}
	if results[checkNext] != CheckStateFail {
		t.Fatalf("next result = %q, want %q", results[checkNext], CheckStateFail)
	}
}

func TestCheckServiceMarksMissingNextConfigNotApplicable(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	service := CheckService{Cache: store.Cache(), SourceClient: &fakeSourceClient{}}
	source := Source{Type: storage.SourceTypeGitHub}
	project := Project{ID: 1, ProviderID: "owner/repo", Name: "Repo"}

	results, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatalf("run project checks: %v", err)
	}
	if results[checkNext] != CheckStateNotApplicable {
		t.Errorf("result %s = %q, want %q", checkNext, results[checkNext], CheckStateNotApplicable)
	}

	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if loaded[checkNext] != CheckStateNotApplicable {
		t.Errorf("loaded %s = %q, want %q", checkNext, loaded[checkNext], CheckStateNotApplicable)
	}
}
