package projectsync

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dnwSilver/tld/internal/storage"
)

type releaseSourceClient struct {
	fakeSourceClient
	tags []Tag
}

func (c *releaseSourceClient) Tags(context.Context, Source, Project) ([]Tag, error) {
	return c.tags, nil
}

func TestReleaseTagClassification(t *testing.T) {
	for _, test := range []struct {
		name string
		kind ReleaseKind
		ok   bool
	}{
		{"v1.2.3", ReleaseKindRelease, true}, {"release/10.20.30", ReleaseKindRelease, true}, {"hotfix/1.2.4", ReleaseKindHotfix, true},
		{"1.2.3", "", false}, {"release/1.2", "", false}, {"v1.2.3-rc.1", "", false}, {"hotfix/a.b.c", "", false}, {"feature/1.2.3", "", false},
	} {
		kind, ok := releaseKindForTag(test.name)
		if kind != test.kind || ok != test.ok {
			t.Errorf("%s: %s %v", test.name, kind, ok)
		}
	}
}

func TestReleaseKindsSurviveCacheAndLegacyDatesLoad(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	date := time.Date(2026, 10, 1, 12, 0, 0, 0, time.FixedZone("test", 5*3600))
	client := &releaseSourceClient{tags: []Tag{{Name: "v1.0.0", CreatedAt: date}, {Name: "release/1.1.0", CreatedAt: date}, {Name: "hotfix/1.1.1", CreatedAt: date}, {Name: "feature/test", CreatedAt: date}, {Name: "v2.0.0"}}}
	service := ReleaseService{Cache: store.Cache(), SourceClient: client}
	source, project := Source{Type: SourceTypeGitLab}, Project{ProviderID: "42"}
	releases, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []Release{{date.UTC(), ReleaseKindRelease}, {date.UTC(), ReleaseKindRelease}, {date.UTC(), ReleaseKindHotfix}}
	if !reflect.DeepEqual(releases, want) {
		t.Fatalf("releases: %#v", releases)
	}
	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil || !reflect.DeepEqual(loaded, want) {
		t.Fatalf("cached: %#v %v", loaded, err)
	}
	legacy, _ := json.Marshal([]time.Time{date.UTC()})
	if err := store.Cache().Set(ctx, CacheNamespaceProjectReleases, releaseSourceCacheKey(source, project.ProviderID), legacy, "application/json", 0); err != nil {
		t.Fatal(err)
	}
	loaded, err = service.LoadProject(ctx, source, project)
	if err != nil || !reflect.DeepEqual(loaded, want[:1]) {
		t.Fatalf("legacy: %#v %v", loaded, err)
	}
}

func TestReleaseStatusRecognizesReleaseAndHotfixVersions(t *testing.T) {
	for _, prefix := range []string{"v", "release/", "hotfix/"} {
		client := &fakeSourceClient{files: map[string][]byte{"package.json": []byte(`{"version":"1.2.3"}`)}}
		statuses := (ReleaseStatusService{SourceClient: client}).computeStatuses(context.Background(), Source{}, Project{StackName: "javascript"}, []Tag{{Name: prefix + "1.2.3"}})
		for _, status := range statuses {
			if status == ReleaseStatusUntaggedMain {
				t.Errorf("%s version marked untagged", prefix)
			}
		}
	}
}
