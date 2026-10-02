package projectsync

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestCheckServiceDoesNotTurnReadFailuresIntoFailedChecks(t *testing.T) {
	denied := errors.New("permission denied")
	source := Source{Type: SourceTypeGitLab}
	project := Project{ProviderID: "42", Name: "repo"}
	for _, tc := range []struct {
		name   string
		client *fakeSourceClient
	}{
		{name: "CI file", client: &fakeSourceClient{fetchErr: denied}},
		{name: "protected branches", client: &fakeSourceClient{protectedBranchesErr: denied}},
		{name: "CI settings", client: &fakeSourceClient{ciSettingsErr: denied}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := CheckService{SourceClient: tc.client}
			if _, err := service.RunProject(context.Background(), source, project, nil); !errors.Is(err, denied) {
				t.Fatalf("RunProject error = %v, want permission denial", err)
			}
		})
	}
}

func TestOperationReadbackRejectsUnconfirmedChange(t *testing.T) {
	var changed atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			changed.Store(true)
			w.Write([]byte(`{}`))
			return
		}
		w.Write([]byte(`{"ci_separated_caches":true}`))
	}))
	defer server.Close()
	source := Source{Type: SourceTypeGitLab, URL: server.URL}
	project := Project{ProviderID: "42"}
	service := OperationService{SourceClient: GitLabClient{HTTPClient: server.Client()}}
	result, err := service.RunAndVerify(context.Background(), source, project, operationShareCICache)
	if err == nil {
		t.Fatal("unconfirmed remote change reported as successful")
	}
	if result.Outcome != OperationOutcomeUncertain {
		t.Fatalf("outcome = %q, want uncertain", result.Outcome)
	}
	if !changed.Load() {
		t.Fatal("operation was not attempted")
	}
}

func TestCISettingsOmittedFieldIsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":42}`))
	}))
	defer server.Close()
	client := GitLabClient{HTTPClient: server.Client()}
	if _, err := client.CISettings(context.Background(), Source{Type: SourceTypeGitLab, URL: server.URL}, Project{ProviderID: "42"}); err == nil {
		t.Fatal("missing ci_separated_caches was treated as false")
	}
}

func TestGitLabOnlyChecksAreNotApplicableToOtherProviders(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}
	for _, definition := range ProjectChecks {
		if !gitlabOnlyCheck(definition.ID) {
			continue
		}
		state, err := service.runCheck(context.Background(), Source{Type: SourceTypeGitHub}, Project{ProviderID: "owner/repo"}, definition.ID, nil, false, nil, false, CISettings{}, false, false, false)
		if err != nil || state != CheckStateNotApplicable {
			t.Errorf("%s = %s, %v", definition.ID, state, err)
		}
	}
}

func TestSourceCacheKeySeparatesOriginAndFullRevision(t *testing.T) {
	a := Source{ID: 1, Type: SourceTypeGitLab, URL: "https://gitlab-a.example"}
	b := Source{ID: 2, Type: SourceTypeGitLab, URL: "https://gitlab-b.example"}
	key := sourceCacheKey(a, "42", "abcdef1234567890", "package.json")
	for _, other := range []string{
		sourceCacheKey(b, "42", "abcdef1234567890", "package.json"),
		sourceCacheKey(Source{ID: 1, Type: SourceTypeGitLab, URL: "https://gitlab-new.example"}, "42", "abcdef1234567890", "package.json"),
		sourceCacheKey(a, "42", "abcdef12ffffffff", "package.json"),
	} {
		if key == other {
			t.Fatalf("cache keys collide: %q", key)
		}
	}
	if strings.Contains(key, "gitlab-a.example") {
		t.Fatal("cache key exposes source URL")
	}
}

func TestVulnerabilityCacheDoesNotCrossSource(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "cache.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := VulnScanService{Cache: store.Cache()}
	a := Source{ID: 1, Type: SourceTypeGitLab, URL: "https://gitlab-a.example"}
	b := Source{ID: 2, Type: SourceTypeGitLab, URL: "https://gitlab-b.example"}
	project := Project{ProviderID: "42"}
	if err := service.cacheReport(ctx, a, project, VulnReport{Scanned: true, Counts: VulnCounts{High: 7}}); err != nil {
		t.Fatal(err)
	}
	if report, err := service.LoadProject(ctx, b, project); err != nil || report.Scanned {
		t.Fatalf("other source inherited report: %#v, %v", report, err)
	}
	if report, err := service.LoadProject(ctx, a, project); err != nil || report.Counts.High != 7 {
		t.Fatalf("source A lost report: %#v, %v", report, err)
	}
}

func TestKotlinMalformedBlockReturnsError(t *testing.T) {
	_, err := (KotlinStrategy{}).Parse("build.gradle", []byte("dependencies {"))
	if err == nil {
		t.Fatal("unterminated dependency block reported success")
	}
}

func FuzzKotlinStrategyDoesNotPanic(f *testing.F) {
	for _, seed := range []string{"dependencies {", "dependencies {\n implementation(\"a:b:1\")\n}", "dependencies { /* } */ implementation 'a:b:1' }"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = (KotlinStrategy{}).Parse("build.gradle", []byte(input))
	})
}

func TestKotlinGroovyMapNotationIsParsed(t *testing.T) {
	deps, err := (KotlinStrategy{}).Parse("build.gradle", []byte("dependencies {\n implementation group: 'org.example', name: 'library', version: '1.0.0'\n}"))
	if err != nil || len(deps) != 1 || deps[0].Name != "org.example:library" || deps[0].Version != "1.0.0" {
		t.Fatalf("map notation: %#v, %v", deps, err)
	}
}

func TestKotlinBracesInStringAndCommentDoNotCloseBlock(t *testing.T) {
	content := []byte("dependencies {\n // }\n def marker = \"{\"\n implementation 'org.example:library:1.0.0'\n}")
	deps, err := (KotlinStrategy{}).Parse("build.gradle", content)
	if err != nil || len(deps) != 1 {
		t.Fatalf("braces in non-code: %#v, %v", deps, err)
	}
}

func TestNpmErrorIsNotCleanReport(t *testing.T) {
	if _, err := parseNpmAudit([]byte(`{"error":{"code":"ENOLOCK"}}`)); err == nil {
		t.Fatal("npm error reported as clean scan")
	}
}

func TestTrivyNoLockfileResultIsNotClean(t *testing.T) {
	for _, output := range [][]byte{[]byte(`{}`), []byte(`{"Results":[]}`), []byte(`{"Results":[{"Target":"README.md"}]}`)} {
		if report, err := parseTrivyReport(output); err == nil || report.Scanned {
			t.Fatalf("accepted incomplete Trivy output %s: %#v, %v", output, report, err)
		}
	}
	report, err := parseTrivyReport([]byte(`{"Results":[{"Target":"Gemfile.lock","Vulnerabilities":[]}]}`))
	if err != nil || !report.Scanned {
		t.Fatalf("rejected clean lockfile scan: %#v, %v", report, err)
	}
}

func TestOsvCLIRequiresEveryParsedDependency(t *testing.T) {
	files := []File{{Path: "build.gradle", Content: []byte("dependencies {\n implementation 'org.a:one:1.0.0'\n implementation 'org.b:two:2.0.0'\n}")}}
	partial := []byte(`{"results":[{"packages":[{"package":{"name":"org.a:one","version":"1.0.0"}}]}]}`)
	if osvCLIHasCoverage(partial, files, KotlinStrategy{}) {
		t.Fatal("partial CLI result reported complete")
	}
	complete := []byte(`{"results":[{"packages":[{"package":{"name":"org.a:one","version":"1.0.0"}},{"package":{"name":"org.b:two","version":"2.0.0"}}]}]}`)
	if !osvCLIHasCoverage(complete, files, KotlinStrategy{}) {
		t.Fatal("complete CLI result rejected")
	}
	if count, ok := osvCLICoverageCount(complete, files, KotlinStrategy{}); !ok || count != 2 {
		t.Fatalf("CLI coverage count = %d, %v", count, ok)
	}
}

func TestGitLabTagsReadSecondPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		count := 100
		if page == "2" {
			count = 1
		}
		items := make([]map[string]any, count)
		for index := range items {
			items[index] = map[string]any{"name": "v1.0.0", "commit": map[string]any{"created_at": "2026-01-01T00:00:00Z"}}
		}
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()
	client, err := NewSourceClient(SourceTypeGitLab, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	tags, err := client.Tags(context.Background(), Source{Type: SourceTypeGitLab, URL: server.URL}, Project{ProviderID: "42"})
	if err != nil || len(tags) != 101 {
		t.Fatalf("tags: %d, %v", len(tags), err)
	}
}

func TestBitbucketTagsRejectCrossOriginNext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[],"next":"https://evil.example/tags"}`))
	}))
	defer server.Close()
	client, err := NewSourceClient(SourceTypeBitbucket, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Tags(context.Background(), Source{Type: SourceTypeBitbucket, URL: server.URL, PATToken: "secret"}, Project{ProviderID: "team/repo"}); err == nil {
		t.Fatal("cross-origin pagination accepted")
	}
}
