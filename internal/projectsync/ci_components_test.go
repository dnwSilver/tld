package projectsync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestCIComponentVersions(t *testing.T) {
	content := []byte(`include:
  - component: $CI_SERVER_FQDN/spectrum-frontend/ci-react-site/build-site@3
  - component: $CI_SERVER_FQDN/spectrum-frontend/ci-react-site/dtrack@3
  - component: $CI_SERVER_FQDN/shared/ci-mr/create-mr@1
  - component: $CI_SERVER_FQDN/shared/ci-ntfy/notify-docker@2
  - component: $CI_SERVER_FQDN/shared/ci-ntfy/notify-store-ios@2
  # - component: $CI_SERVER_FQDN/shared/ci-ntfy/notify@main
`)

	versions := ciComponentVersions(content)
	want := map[string]string{
		checkCICD:   "3",
		checkDtrack: "3",
		checkCremr:  "1",
		checkNtfy:   "2",
	}
	for checkID, expected := range want {
		if versions[checkID] != expected {
			t.Errorf("version %s = %q, want %q", checkID, versions[checkID], expected)
		}
	}
}

func TestCIComponentVersionsSupportsLegacyTemplateRef(t *testing.T) {
	content := []byte(`include:
  - project: spectrum-frontend/ci-templates
    ref: master
    file: /templates/.gitlab-ci.yml
`)

	versions := ciComponentVersions(content)
	if versions[checkCICD] != "master" {
		t.Fatalf("ci/cd version = %q, want master", versions[checkCICD])
	}
}

func TestCheckServiceCachesComponentVersion(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	client := &fakeSourceClient{files: map[string][]byte{
		gitlabCIFile: []byte(`include:
  - component: $CI_SERVER_FQDN/spectrum-frontend/ci-react-site/dtrack@3
`),
	}}
	service := CheckService{Cache: store.Cache(), SourceClient: client}
	source := Source{Type: storage.SourceTypeGitLab}
	project := Project{ID: 1, ProviderID: "group/repo", Name: "Repo"}
	if _, err := service.RunProject(ctx, source, project, nil); err != nil {
		t.Fatalf("run project checks: %v", err)
	}

	results, versions, err := service.LoadProjectWithVersions(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if results[checkDtrack] != CheckStatePass {
		t.Fatalf("dtrack state = %q, want %q", results[checkDtrack], CheckStatePass)
	}
	if versions[checkDtrack] != "3" {
		t.Fatalf("dtrack version = %q, want 3", versions[checkDtrack])
	}
}
