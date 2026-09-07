package projectsync

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestJavaScriptStrategyParsesDependencyGroups(t *testing.T) {
	strategy := JavaScriptStrategy{}
	dependencies, err := strategy.Parse("package.json", []byte(`{
		"dependencies": {"react": "^19.0.0"},
		"devDependencies": {"typescript": "^5.0.0"},
		"peerDependencies": {"next": "^15.0.0"},
		"optionalDependencies": {"sharp": "^0.33.0"},
		"engines": {"node": ">=22.0.0"}
	}`))
	if err != nil {
		t.Fatalf("parse package.json: %v", err)
	}
	if len(dependencies) != 5 {
		t.Fatalf("len(dependencies) = %d, want 5", len(dependencies))
	}

	types := map[string]string{}
	versions := map[string]string{}
	for _, dependency := range dependencies {
		types[dependency.Name] = dependency.DependencyType
		versions[dependency.Name] = dependency.Version
	}
	if types["react"] != DependencyTypeRuntime || types["typescript"] != DependencyTypeDev || types["next"] != DependencyTypePeer || types["sharp"] != DependencyTypeOptional || types["node"] != DependencyTypeEngines {
		t.Fatalf("dependency types = %#v", types)
	}
	if versions["node"] != ">=22.0.0" {
		t.Fatalf("node version = %q, want >=22.0.0", versions["node"])
	}

	lockDependencies, err := strategy.Parse("package-lock.json", []byte(`{"lockfileVersion": 3}`))
	if err != nil {
		t.Fatalf("parse package-lock.json: %v", err)
	}
	if len(lockDependencies) != 0 {
		t.Fatalf("lock dependencies = %d, want 0", len(lockDependencies))
	}
}

func TestJavaScriptStrategyParsesNvmrc(t *testing.T) {
	strategy := JavaScriptStrategy{}

	dependencies, err := strategy.Parse(".nvmrc", []byte("22.11.0\n"))
	if err != nil {
		t.Fatalf("parse .nvmrc: %v", err)
	}
	if len(dependencies) != 1 {
		t.Fatalf("len(dependencies) = %d, want 1", len(dependencies))
	}
	if dependencies[0].Name != "node" || dependencies[0].Version != "22.11.0" || dependencies[0].DependencyType != DependencyTypeNvmrc {
		t.Fatalf("dependency = %#v", dependencies[0])
	}

	dependencies, err = strategy.Parse(".nvmrc", []byte("# comment\nv20.18.0\n"))
	if err != nil {
		t.Fatalf("parse .nvmrc with prefix: %v", err)
	}
	if len(dependencies) != 1 || dependencies[0].Version != "20.18.0" {
		t.Fatalf("dependency = %#v", dependencies)
	}

	dependencies, err = strategy.Parse(".nvmrc", []byte("# only comment\n"))
	if err != nil {
		t.Fatalf("parse empty .nvmrc: %v", err)
	}
	if len(dependencies) != 0 {
		t.Fatalf("len(dependencies) = %d, want 0", len(dependencies))
	}
}

func TestPreferNodeFromNvmrc(t *testing.T) {
	dependencies := preferNodeFromNvmrc([]storage.ProjectDependency{
		{Name: "node", Version: ">=22.0.0", DependencyType: DependencyTypeEngines, SourceFile: "package.json"},
		{Name: "node", Version: "22.11.0", DependencyType: DependencyTypeNvmrc, SourceFile: ".nvmrc"},
		{Name: "react", Version: "^19.0.0", DependencyType: DependencyTypeRuntime, SourceFile: "package.json"},
	})
	if len(dependencies) != 2 {
		t.Fatalf("len(dependencies) = %d, want 2", len(dependencies))
	}
	if dependencies[0].Name != "node" || dependencies[0].Version != "22.11.0" || dependencies[0].SourceFile != ".nvmrc" {
		t.Fatalf("node dependency = %#v", dependencies[0])
	}
	if dependencies[1].Name != "react" {
		t.Fatalf("react dependency = %#v", dependencies[1])
	}
}

func TestSwiftStrategyParsesLockedVersions(t *testing.T) {
	strategy := SwiftStrategy{}
	gems, err := strategy.Parse("Gemfile.lock", []byte(`GEM
  remote: https://rubygems.org/
  specs:
    fastlane (2.227.2)
      addressable (>= 2.8, < 3.0.0)
    xcodeproj (1.27.0)

PLATFORMS
  ruby
`))
	if err != nil {
		t.Fatalf("parse Gemfile.lock: %v", err)
	}
	pods, err := strategy.Parse("Podfile.lock", []byte(`PODS:
  - Alamofire (5.10.2)
  - Firebase/CoreOnly (11.15.0):
    - FirebaseCore (= 11.15.0)
  - FirebaseCore (11.15.0)

DEPENDENCIES:
  - Alamofire
`))
	if err != nil {
		t.Fatalf("parse Podfile.lock: %v", err)
	}

	dependencies := append(gems, pods...)
	values := map[string]string{}
	for _, dependency := range dependencies {
		values[dependency.Name] = dependency.Version
	}
	if values["fastlane"] != "2.227.2" || values["xcodeproj"] != "1.27.0" {
		t.Fatalf("gem dependencies = %#v", values)
	}
	if values["Alamofire"] != "5.10.2" || values["Firebase/CoreOnly"] != "11.15.0" || values["FirebaseCore"] != "11.15.0" {
		t.Fatalf("pod dependencies = %#v", values)
	}
}

func TestKotlinStrategyParsesVersionCatalog(t *testing.T) {
	strategy := KotlinStrategy{}
	dependencies, err := strategy.Parse("gradle/libs.versions.toml", []byte(`[versions]
coreKtx = "1.15.0"
okhttp = "4.12.0"
agp = "8.6.1"

[libraries]
androidx-core-ktx = { group = "androidx.core", name = "core-ktx", version.ref = "coreKtx" }
agcp = { module = "com.huawei.agconnect:agcp", version = "1.9.1.301" }
androidx-ui = { group = "androidx.compose.ui", name = "ui" }
okhttp = { group = "com.squareup.okhttp3", name = "okhttp", version.ref = "okhttp" }

[plugins]
android-application = { id = "com.android.application", version.ref = "agp" }
`))
	if err != nil {
		t.Fatalf("parse libs.versions.toml: %v", err)
	}
	if len(dependencies) != 4 {
		t.Fatalf("len(dependencies) = %d, want 4", len(dependencies))
	}

	versions := map[string]string{}
	types := map[string]string{}
	for _, dependency := range dependencies {
		versions[dependency.Name] = dependency.Version
		types[dependency.Name] = dependency.DependencyType
	}
	if versions["androidx.core:core-ktx"] != "1.15.0" {
		t.Fatalf("core-ktx version = %q, want 1.15.0", versions["androidx.core:core-ktx"])
	}
	if versions["com.huawei.agconnect:agcp"] != "1.9.1.301" {
		t.Fatalf("agcp version = %q, want 1.9.1.301", versions["com.huawei.agconnect:agcp"])
	}
	if versions["androidx.compose.ui:ui"] != "" {
		t.Fatalf("compose ui version = %q, want empty", versions["androidx.compose.ui:ui"])
	}
	if types["androidx.core:core-ktx"] != DependencyTypeGradleLibrary {
		t.Fatalf("core-ktx type = %q, want %q", types["androidx.core:core-ktx"], DependencyTypeGradleLibrary)
	}
	if types["com.android.application"] != DependencyTypeGradlePlugin || versions["com.android.application"] != "8.6.1" {
		t.Fatalf("plugin = %q %q", types["com.android.application"], versions["com.android.application"])
	}
}

func TestKotlinStrategyParsesInlineGradleDependencies(t *testing.T) {
	strategy := KotlinStrategy{}
	dependencies, err := strategy.Parse("app/build.gradle.kts", []byte(`val composeVersion = "1.6.8"

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation(platform("androidx.compose:compose-bom:2024.06.00"))
    implementation("androidx.compose.ui:ui:$composeVersion")
    implementation("com.google.firebase:firebase-messaging-ktx")
    ksp("com.google.dagger:hilt-android-compiler:2.51.1")
    "huaweiImplementation"("com.huawei.hms:push:6.11.0.300")
    testImplementation("junit:junit:4.13.2")
}
`))
	if err != nil {
		t.Fatalf("parse build.gradle.kts: %v", err)
	}

	versions := map[string]string{}
	types := map[string]string{}
	for _, dependency := range dependencies {
		versions[dependency.Name] = dependency.Version
		types[dependency.Name] = dependency.DependencyType
	}
	if versions["androidx.core:core-ktx"] != "1.13.1" {
		t.Fatalf("core-ktx version = %q, want 1.13.1", versions["androidx.core:core-ktx"])
	}
	if versions["androidx.compose:compose-bom"] != "2024.06.00" {
		t.Fatalf("compose-bom version = %q", versions["androidx.compose:compose-bom"])
	}
	if versions["androidx.compose.ui:ui"] != "1.6.8" {
		t.Fatalf("compose ui version = %q, want 1.6.8 (resolved val)", versions["androidx.compose.ui:ui"])
	}
	if _, ok := versions["com.google.firebase:firebase-messaging-ktx"]; ok {
		t.Fatal("BOM-managed dependency without version must be skipped")
	}
	if types["com.huawei.hms:push"] != "huaweiImplementation" {
		t.Fatalf("hms push type = %q", types["com.huawei.hms:push"])
	}
	if types["junit:junit"] != "testImplementation" || versions["junit:junit"] != "4.13.2" {
		t.Fatalf("junit = %q %q", types["junit:junit"], versions["junit:junit"])
	}
}

func TestKotlinStrategyParsesGroovyGradleDependencies(t *testing.T) {
	strategy := KotlinStrategy{}
	dependencies, err := strategy.Parse("app/build.gradle", []byte(`dependencies {
    implementation 'androidx.core:core-ktx:1.12.0'
    implementation platform('androidx.compose:compose-bom:2022.10.00')
    implementation 'androidx.compose.ui:ui'
    implementation "androidx.appcompat:appcompat:1.6.1"
    implementation("com.jakewharton.timber:timber:5.0.1")
    testImplementation 'junit:junit:4.13.2'
    debugImplementation 'androidx.compose.ui:ui-tooling'
}
`))
	if err != nil {
		t.Fatalf("parse build.gradle: %v", err)
	}

	versions := map[string]string{}
	types := map[string]string{}
	for _, dependency := range dependencies {
		versions[dependency.Name] = dependency.Version
		types[dependency.Name] = dependency.DependencyType
	}
	if versions["androidx.core:core-ktx"] != "1.12.0" {
		t.Fatalf("core-ktx version = %q, want 1.12.0", versions["androidx.core:core-ktx"])
	}
	if versions["androidx.compose:compose-bom"] != "2022.10.00" {
		t.Fatalf("compose-bom version = %q", versions["androidx.compose:compose-bom"])
	}
	if versions["androidx.appcompat:appcompat"] != "1.6.1" {
		t.Fatalf("appcompat version = %q, want 1.6.1", versions["androidx.appcompat:appcompat"])
	}
	if _, ok := versions["androidx.compose.ui:ui"]; ok {
		t.Fatal("BOM-managed dependency without version must be skipped")
	}
	if types["junit:junit"] != "testImplementation" || versions["junit:junit"] != "4.13.2" {
		t.Fatalf("junit = %q %q", types["junit:junit"], versions["junit:junit"])
	}
	if types["com.jakewharton.timber:timber"] != "implementation" {
		t.Fatalf("timber type = %q", types["com.jakewharton.timber:timber"])
	}
}

func TestGoStrategyParsesGoModRequires(t *testing.T) {
	strategy := GoStrategy{}
	dependencies, err := strategy.Parse("go.mod", []byte(`module github.com/dnwSilver/tld

go 1.24.1

require github.com/charmbracelet/lipgloss v1.1.0

require (
	github.com/charmbracelet/bubbletea v1.3.6
	golang.org/x/term v0.33.0 // indirect
)
`))
	if err != nil {
		t.Fatalf("parse go.mod: %v", err)
	}
	if len(dependencies) != 3 {
		t.Fatalf("len(dependencies) = %d, want 3", len(dependencies))
	}

	values := map[string]string{}
	types := map[string]string{}
	for _, dependency := range dependencies {
		values[dependency.Name] = dependency.Version
		types[dependency.Name] = dependency.DependencyType
	}
	if values["github.com/charmbracelet/lipgloss"] != "v1.1.0" || values["github.com/charmbracelet/bubbletea"] != "v1.3.6" || values["golang.org/x/term"] != "v0.33.0" {
		t.Fatalf("go dependencies = %#v", values)
	}
	if types["github.com/charmbracelet/lipgloss"] != DependencyTypeGoModule {
		t.Fatalf("go dependency types = %#v", types)
	}

	sumDependencies, err := strategy.Parse("go.sum", []byte(`github.com/charmbracelet/lipgloss v1.1.0 h1:abc`))
	if err != nil {
		t.Fatalf("parse go.sum: %v", err)
	}
	if len(sumDependencies) != 0 {
		t.Fatalf("go.sum dependencies = %d, want 0", len(sumDependencies))
	}
}

func TestServiceCachesSwiftLocks(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	stack, err := store.Stacks().Create(ctx, "", "Swift", "#F05138")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	sourceEntity, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", storage.SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	projectEntity, err := store.Projects().Create(ctx, "owner/ios", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "iOS", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	client := &fakeSourceClient{
		files: map[string][]byte{
			"Gemfile.lock": []byte(`GEM
  specs:
    fastlane (2.227.2)
`),
			"Podfile.lock": []byte(`PODS:
  - Alamofire (5.10.2)
`),
		},
	}
	service := Service{
		Cache:        store.Cache(),
		Runs:         store.ProjectDependencies(),
		SourceClient: client,
	}
	result, err := service.Sync(ctx, Source{Type: storage.SourceTypeGitHub}, Project{
		ID:         projectEntity.ID,
		ProviderID: projectEntity.ProjectID,
		Name:       projectEntity.Name,
		StackName:  stack.Name,
	}, nil)
	if err != nil {
		t.Fatalf("sync swift project: %v", err)
	}
	if result.Count != 2 {
		t.Fatalf("count = %d, want 2", result.Count)
	}

	if client.fetches["Gemfile.lock"] != 1 || client.fetches["Podfile.lock"] != 1 {
		t.Fatalf("fetches = %#v", client.fetches)
	}
	if _, err := store.Cache().Get(ctx, CacheNamespaceProjectFiles, "github:owner/ios:abcdef12:Gemfile.lock"); err != nil {
		t.Fatalf("get Gemfile.lock cache: %v", err)
	}
	if _, err := store.Cache().Get(ctx, CacheNamespaceProjectFiles, "github:owner/ios:abcdef12:Podfile.lock"); err != nil {
		t.Fatalf("get Podfile.lock cache: %v", err)
	}
}

func TestServiceCachesGoSum(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	stack, err := store.Stacks().Create(ctx, "", "Golang", "#00ADD8")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	sourceEntity, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", storage.SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	projectEntity, err := store.Projects().Create(ctx, "owner/go", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "Go App", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	client := &fakeSourceClient{
		files: map[string][]byte{
			"go.mod": []byte(`module github.com/example/app

go 1.24.1

require github.com/charmbracelet/lipgloss v1.1.0
`),
			"go.sum": []byte(`github.com/charmbracelet/lipgloss v1.1.0 h1:abc`),
		},
	}
	service := Service{
		Cache:        store.Cache(),
		Runs:         store.ProjectDependencies(),
		SourceClient: client,
	}
	result, err := service.Sync(ctx, Source{Type: storage.SourceTypeGitHub}, Project{
		ID:         projectEntity.ID,
		ProviderID: projectEntity.ProjectID,
		Name:       projectEntity.Name,
		StackName:  stack.Name,
	}, nil)
	if err != nil {
		t.Fatalf("sync go project: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("count = %d, want 1", result.Count)
	}
	if client.fetches["go.mod"] != 1 || client.fetches["go.sum"] != 1 {
		t.Fatalf("fetches = %#v", client.fetches)
	}
	if _, err := store.Cache().Get(ctx, CacheNamespaceProjectFiles, "github:owner/go:abcdef12:go.sum"); err != nil {
		t.Fatalf("get go.sum cache: %v", err)
	}
}

func TestServiceCachesPackageLock(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	stack, err := store.Stacks().Create(ctx, "", "JavaScript", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	sourceEntity, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", storage.SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	projectEntity, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	client := &fakeSourceClient{
		files: map[string][]byte{
			"package.json":      []byte(`{"dependencies":{"react":"^19.0.0"}}`),
			"package-lock.json": []byte(`{"lockfileVersion":3}`),
		},
	}
	service := Service{
		Cache:        store.Cache(),
		Runs:         store.ProjectDependencies(),
		SourceClient: client,
	}
	result, err := service.Sync(ctx, Source{Type: storage.SourceTypeGitHub}, Project{
		ID:         projectEntity.ID,
		ProviderID: projectEntity.ProjectID,
		Name:       projectEntity.Name,
		StackName:  stack.Name,
	}, nil)
	if err != nil {
		t.Fatalf("sync project: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("count = %d, want 1", result.Count)
	}
	if client.fetches["package-lock.json"] != 1 {
		t.Fatalf("package-lock fetches = %d, want 1", client.fetches["package-lock.json"])
	}

	entry, err := store.Cache().Get(ctx, CacheNamespaceProjectFiles, "github:owner/repo:abcdef12:package-lock.json")
	if err != nil {
		t.Fatalf("get package-lock cache: %v", err)
	}
	if string(entry.Value) != `{"lockfileVersion":3}` {
		t.Fatalf("package-lock cache = %q", string(entry.Value))
	}

	result, err = service.Sync(ctx, Source{Type: storage.SourceTypeGitHub}, Project{
		ID:         projectEntity.ID,
		ProviderID: projectEntity.ProjectID,
		Name:       projectEntity.Name,
		StackName:  stack.Name,
	}, nil)
	if err != nil {
		t.Fatalf("sync project again: %v", err)
	}
	if !result.UpToDate {
		t.Fatal("expected up to date result")
	}
}

func TestReleaseStatusUsesDevBranchDivergence(t *testing.T) {
	client := &fakeSourceClient{
		files: map[string][]byte{
			"package.json": []byte(`{"version":"1.2.3"}`),
		},
		divergence: BranchDivergence{Ahead: 2, Behind: 1},
	}
	service := ReleaseStatusService{SourceClient: client}

	statuses := service.computeStatuses(context.Background(), Source{Type: storage.SourceTypeGitHub}, Project{
		ProviderID: "owner/repo",
		Name:       "Repo",
		StackName:  "JavaScript",
	}, []Tag{{Name: "v1.2.3"}})

	if len(statuses) != 2 {
		t.Fatalf("statuses = %#v, want 2 statuses", statuses)
	}
	if statuses[0] != ReleaseStatusDevAhead || statuses[1] != ReleaseStatusDevBehind {
		t.Fatalf("statuses = %#v, want dev ahead and dev behind", statuses)
	}
}

func TestGitHubClientUsesInjectedHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/repos/owner/repo":
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		case "/api/v3/repos/owner/repo/commits/main":
			_, _ = w.Write([]byte(`{"sha":"abcdef1234567890"}`))
		case "/api/v3/repos/owner/repo/contents/package.json":
			if r.URL.Query().Get("ref") != "abcdef1234567890" {
				t.Fatalf("ref = %q", r.URL.Query().Get("ref"))
			}
			encoded := base64.StdEncoding.EncodeToString([]byte(`{"dependencies":{}}`))
			_, _ = w.Write([]byte(`{"encoding":"base64","content":"` + encoded + `"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := GitHubClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitHub, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "owner/repo"}
	commit, err := client.ResolveHead(context.Background(), source, project)
	if err != nil {
		t.Fatalf("resolve head: %v", err)
	}
	if commit.ShortSHA != "abcdef12" {
		t.Fatalf("short sha = %q, want abcdef12", commit.ShortSHA)
	}

	content, err := client.FetchFile(context.Background(), source, project, commit.SHA, "package.json")
	if err != nil {
		t.Fatalf("fetch file: %v", err)
	}
	if string(content) != `{"dependencies":{}}` {
		t.Fatalf("content = %q", string(content))
	}
}

func TestGitHubClientHasBranchAndDefaultBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/repos/owner/repo":
			_, _ = w.Write([]byte(`{"default_branch":"dev"}`))
		case "/api/v3/repos/owner/repo/branches/master":
			_, _ = w.Write([]byte(`{"name":"master"}`))
		case "/api/v3/repos/owner/repo/branches/dev":
			_, _ = w.Write([]byte(`{"name":"dev"}`))
		case "/api/v3/repos/owner/repo/branches/missing":
			http.NotFound(w, r)
		case "/api/v3/repos/owner/repo/compare/master...dev":
			_, _ = w.Write([]byte(`{"ahead_by":5,"behind_by":2}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := GitHubClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitHub, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "owner/repo"}

	defaultBranch, err := client.DefaultBranch(context.Background(), source, project)
	if err != nil {
		t.Fatalf("default branch: %v", err)
	}
	if defaultBranch != "dev" {
		t.Fatalf("default branch = %q, want dev", defaultBranch)
	}

	hasMaster, err := client.HasBranch(context.Background(), source, project, "master")
	if err != nil || !hasMaster {
		t.Fatalf("has master = %v, err = %v, want true", hasMaster, err)
	}
	hasMissing, err := client.HasBranch(context.Background(), source, project, "missing")
	if err != nil || hasMissing {
		t.Fatalf("has missing = %v, err = %v, want false", hasMissing, err)
	}
	divergence, err := client.CompareBranches(context.Background(), source, project, "master", "dev")
	if err != nil {
		t.Fatalf("compare branches: %v", err)
	}
	if divergence.Ahead != 5 || divergence.Behind != 2 {
		t.Fatalf("divergence = %#v, want ahead 5 behind 2", divergence)
	}
}

func TestGitLabClientHasBranchAndDefaultBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/api/v4/projects/group%2Frepo":
			_, _ = w.Write([]byte(`{"default_branch":"dev"}`))
		case "/api/v4/projects/group%2Frepo/repository/branches/master":
			_, _ = w.Write([]byte(`{"name":"master"}`))
		case "/api/v4/projects/group%2Frepo/repository/branches/missing":
			http.NotFound(w, r)
		case "/api/v4/projects/group%2Frepo/repository/compare":
			switch r.URL.Query().Get("from") + "..." + r.URL.Query().Get("to") {
			case "master...dev":
				_, _ = w.Write([]byte(`{"commits":[{}, {}, {}]}`))
			case "dev...master":
				_, _ = w.Write([]byte(`{"commits":[{}]}`))
			default:
				http.NotFound(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := GitLabClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	defaultBranch, err := client.DefaultBranch(context.Background(), source, project)
	if err != nil {
		t.Fatalf("default branch: %v", err)
	}
	if defaultBranch != "dev" {
		t.Fatalf("default branch = %q, want dev", defaultBranch)
	}

	hasMaster, err := client.HasBranch(context.Background(), source, project, "master")
	if err != nil || !hasMaster {
		t.Fatalf("has master = %v, err = %v, want true", hasMaster, err)
	}
	hasMissing, err := client.HasBranch(context.Background(), source, project, "missing")
	if err != nil || hasMissing {
		t.Fatalf("has missing = %v, err = %v, want false", hasMissing, err)
	}
	divergence, err := client.CompareBranches(context.Background(), source, project, "master", "dev")
	if err != nil {
		t.Fatalf("compare branches: %v", err)
	}
	if divergence.Ahead != 3 || divergence.Behind != 1 {
		t.Fatalf("divergence = %#v, want ahead 3 behind 1", divergence)
	}
}

func TestGitLabClientListsPipelineSchedules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v4/projects/group%2Frepo/pipeline_schedules" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("PRIVATE-TOKEN") != "secret" {
			t.Fatalf("PRIVATE-TOKEN = %q, want secret", r.Header.Get("PRIVATE-TOKEN"))
		}
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("per_page") != "100" {
			t.Fatalf("pagination query = %q, want page=1 and per_page=100", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[{
			"description":"🌚 Nightly build",
			"ref":"refs/heads/master",
			"owner":{"username":"group349_bot2"}
		}]`))
	}))
	defer server.Close()

	client := GitLabClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	schedules, err := client.PipelineSchedules(context.Background(), source, project)
	if err != nil {
		t.Fatalf("list pipeline schedules: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("pipeline schedules = %d, want 1", len(schedules))
	}
	if schedules[0].Description != nightlyScheduleDescription ||
		schedules[0].Ref != "refs/heads/master" ||
		schedules[0].OwnerUsername != nightlyScheduleOwner {
		t.Fatalf("pipeline schedule = %#v", schedules[0])
	}
}

func TestNightlyScheduleCheckPassesWhenConfigured(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{
		schedules: []PipelineSchedule{{
			Description:   nightlyScheduleDescription,
			Ref:           "refs/heads/" + nightlyScheduleRef,
			OwnerUsername: "@" + nightlyScheduleOwner,
		}},
	}}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkNightly,
		nil,
		false,
		nil,
		false,
		CISettings{},
		false,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run nightly check: %v", err)
	}
	if state != CheckStatePass {
		t.Fatalf("nightly check state = %q, want %q", state, CheckStatePass)
	}
}

func TestNightlyScheduleCheckWarnsWhenOwnerIsMisconfigured(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{
		schedules: []PipelineSchedule{{
			Description:   nightlyScheduleDescription,
			Ref:           "refs/heads/" + nightlyScheduleRef,
			OwnerUsername: "another-user",
		}},
	}}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkNightly,
		nil,
		false,
		nil,
		false,
		CISettings{},
		false,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run nightly check: %v", err)
	}
	if state != CheckStateWarning {
		t.Fatalf("nightly check state = %q, want %q", state, CheckStateWarning)
	}
}

func TestNightlyScheduleCheckFailsWhenMissing(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkNightly,
		nil,
		false,
		nil,
		false,
		CISettings{},
		false,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run nightly check: %v", err)
	}
	if state != CheckStateFail {
		t.Fatalf("nightly check state = %q, want %q", state, CheckStateFail)
	}
}

func TestSeparatedCachesCheckPassesWhenConfigured(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}
	settings := CISettings{SeparatedCaches: false}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkSeparatedCaches,
		nil,
		false,
		nil,
		false,
		settings,
		true,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run %s check: %v", checkSeparatedCaches, err)
	}
	if state != CheckStatePass {
		t.Fatalf("%s check state = %q, want %q", checkSeparatedCaches, state, CheckStatePass)
	}
}

func TestSeparatedCachesCheckFailsWhenMisconfigured(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}
	settings := CISettings{SeparatedCaches: true}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkSeparatedCaches,
		nil,
		false,
		nil,
		false,
		settings,
		true,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run %s check: %v", checkSeparatedCaches, err)
	}
	if state != CheckStateFail {
		t.Fatalf("%s check state = %q, want %q", checkSeparatedCaches, state, CheckStateFail)
	}
}

func TestSeparatedCachesCheckFailsWhenSettingsAreUnavailable(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}

	state, err := service.runCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
		checkSeparatedCaches,
		nil,
		false,
		nil,
		false,
		CISettings{},
		false,
		nil,
		false,
		false,
	)
	if err != nil {
		t.Fatalf("run %s check: %v", checkSeparatedCaches, err)
	}
	if state != CheckStateFail {
		t.Fatalf("%s check state = %q, want %q", checkSeparatedCaches, state, CheckStateFail)
	}
}

func TestProcessModeCheckUsesReleaseCandidateResourceGroup(t *testing.T) {
	client := &fakeSourceClient{
		numericProjectID: 1986,
		resourceGroupModes: map[string]string{
			"application-rc-1986": processModeOldestFirst,
		},
	}
	service := CheckService{SourceClient: client}

	state, err := service.runProcessModeCheck(
		context.Background(),
		Source{Type: storage.SourceTypeGitLab},
		Project{ProviderID: "group/repo"},
	)
	if err != nil {
		t.Fatalf("run process mode check: %v", err)
	}
	if state != CheckStatePass {
		t.Fatalf("process mode state = %q, want %q", state, CheckStatePass)
	}
	if client.requestedResourceGroup != "application-rc-1986" {
		t.Fatalf("requested resource group = %q, want application-rc-1986", client.requestedResourceGroup)
	}
}

func TestProcessModeCheckFailsForWrongOrMissingResourceGroup(t *testing.T) {
	tests := []struct {
		name  string
		modes map[string]string
	}{
		{name: "wrong mode", modes: map[string]string{"application-rc-7": "unordered"}},
		{name: "missing group"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := CheckService{SourceClient: &fakeSourceClient{resourceGroupModes: test.modes}}
			state, err := service.runProcessModeCheck(
				context.Background(),
				Source{Type: storage.SourceTypeGitLab},
				Project{ProviderID: "7"},
			)
			if err != nil {
				t.Fatalf("run process mode check: %v", err)
			}
			if state != CheckStateFail {
				t.Fatalf("process mode state = %q, want %q", state, CheckStateFail)
			}
		})
	}
}

func TestGitLabSettingsChecksNotApplicableForNonGitLabSources(t *testing.T) {
	service := CheckService{SourceClient: &fakeSourceClient{}}

	for _, checkID := range []string{checkProcessMode, checkSeparatedCaches} {
		state, err := service.runCheck(
			context.Background(),
			Source{Type: storage.SourceTypeGitHub},
			Project{ProviderID: "owner/repo"},
			checkID,
			nil,
			false,
			nil,
			false,
			CISettings{},
			false,
			nil,
			false,
			false,
		)
		if err != nil {
			t.Fatalf("run %s check: %v", checkID, err)
		}
		if state != CheckStateNotApplicable {
			t.Fatalf("%s check state = %q, want %q", checkID, state, CheckStateNotApplicable)
		}
	}
}

func TestCheckServiceCachesResults(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	client := &fakeSourceClient{}
	service := CheckService{
		Cache:        store.Cache(),
		SourceClient: client,
	}
	source := Source{Type: storage.SourceTypeGitHub}
	project := Project{ID: 1, ProviderID: "owner/repo", Name: "Repo"}

	results, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatalf("run project checks: %v", err)
	}
	if results["master"] != CheckStatePass || results["dev"] != CheckStatePass || results["default"] != CheckStatePass {
		t.Fatalf("results = %#v", results)
	}

	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if loaded["master"] != CheckStatePass || loaded["dev"] != CheckStatePass || loaded["default"] != CheckStatePass {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestCheckServiceCachesWarningState(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	service := CheckService{Cache: store.Cache()}
	source := Source{Type: storage.SourceTypeGitLab}
	project := Project{ProviderID: "group/repo"}
	if err := service.cacheResult(ctx, source, project, checkNightly, CheckStateWarning, ""); err != nil {
		t.Fatalf("cache warning state: %v", err)
	}

	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if loaded[checkNightly] != CheckStateWarning {
		t.Fatalf("loaded nightly state = %q, want %q", loaded[checkNightly], CheckStateWarning)
	}
}

func TestGitLabClientReadsCISettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v4/projects/group%2Frepo" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("PRIVATE-TOKEN") != "secret" {
			t.Fatalf("PRIVATE-TOKEN = %q, want secret", r.Header.Get("PRIVATE-TOKEN"))
		}
		_, _ = w.Write([]byte(`{
			"ci_separated_caches": true
		}`))
	}))
	defer server.Close()

	client := GitLabClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	settings, err := client.CISettings(context.Background(), source, project)
	if err != nil {
		t.Fatalf("read ci settings: %v", err)
	}
	if !settings.SeparatedCaches {
		t.Fatal("separated caches = false, want true")
	}
}

func TestGitLabClientReadsResourceGroupProcessMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v4/projects/group%2Frepo/resource_groups/application-rc-1986" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("PRIVATE-TOKEN") != "secret" {
			t.Fatalf("PRIVATE-TOKEN = %q, want secret", r.Header.Get("PRIVATE-TOKEN"))
		}
		_, _ = w.Write([]byte(`{"key":"application-rc-1986","process_mode":"oldest_first"}`))
	}))
	defer server.Close()

	client := GitLabClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	processMode, err := client.ResourceGroupProcessMode(context.Background(), source, project, "application-rc-1986")
	if err != nil {
		t.Fatalf("read resource group process mode: %v", err)
	}
	if processMode != processModeOldestFirst {
		t.Fatalf("process mode = %q, want %q", processMode, processModeOldestFirst)
	}
}

func TestCheckServiceRunsCIAndResourceGroupSettingsChecks(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	client := &fakeSourceClient{
		ciSettings:       CISettings{SeparatedCaches: false},
		numericProjectID: 1986,
		resourceGroupModes: map[string]string{
			"application-rc-1986": processModeOldestFirst,
		},
	}
	service := CheckService{
		Cache:        store.Cache(),
		SourceClient: client,
	}
	source := Source{Type: storage.SourceTypeGitLab}
	project := Project{ID: 1, ProviderID: "group/repo", Name: "Repo"}

	results, err := service.RunProject(ctx, source, project, nil)
	if err != nil {
		t.Fatalf("run project checks: %v", err)
	}
	if results[checkProcessMode] != CheckStatePass || results[checkSeparatedCaches] != CheckStatePass {
		t.Fatalf("results = %#v", results)
	}

	loaded, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load project checks: %v", err)
	}
	if loaded[checkProcessMode] != CheckStatePass || loaded[checkSeparatedCaches] != CheckStatePass {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestGitLabClientHasMaintainerRights(t *testing.T) {
	cases := []struct {
		name        string
		permissions string
		want        bool
	}{
		{name: "maintainer via group", permissions: `{"project_access": null, "group_access": {"access_level": 40}}`, want: true},
		{name: "owner via project", permissions: `{"project_access": {"access_level": 50}, "group_access": null}`, want: true},
		{name: "developer only", permissions: `{"project_access": {"access_level": 30}, "group_access": null}`, want: false},
		{name: "no access info", permissions: `{"project_access": null, "group_access": null}`, want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.EscapedPath() != "/api/v4/projects/group%2Frepo" {
					http.NotFound(w, r)
					return
				}
				_, _ = w.Write([]byte(`{"id": 1, "permissions": ` + testCase.permissions + `}`))
			}))
			defer server.Close()

			client := GitLabClient{HTTPClient: server.Client()}
			source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
			project := Project{ProviderID: "group/repo"}

			maintainer, err := client.HasMaintainerRights(context.Background(), source, project)
			if err != nil {
				t.Fatalf("has maintainer rights: %v", err)
			}
			if maintainer != testCase.want {
				t.Fatalf("maintainer = %v, want %v", maintainer, testCase.want)
			}
		})
	}
}

func TestOperationServiceSharesCICache(t *testing.T) {
	var method, path, body, token string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		method = r.Method
		path = r.URL.EscapedPath()
		body = string(payload)
		token = r.Header.Get("PRIVATE-TOKEN")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	service := OperationService{SourceClient: GitLabClient{HTTPClient: server.Client()}}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	if err := service.Run(context.Background(), source, project, operationShareCICache); err != nil {
		t.Fatalf("run share ci cache: %v", err)
	}
	if method != http.MethodPut {
		t.Fatalf("method = %q, want PUT", method)
	}
	if path != "/api/v4/projects/group%2Frepo" {
		t.Fatalf("path = %q", path)
	}
	if body != "ci_separated_caches=false" {
		t.Fatalf("body = %q", body)
	}
	if token != "secret" {
		t.Fatalf("PRIVATE-TOKEN = %q, want secret", token)
	}
}

func TestOperationServiceSetsResourceGroupProcessMode(t *testing.T) {
	var putBody string
	putCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/api/v4/projects/group%2Frepo":
			_, _ = w.Write([]byte(`{"id": 1986}`))
		case r.Method == http.MethodPut && r.URL.EscapedPath() == "/api/v4/projects/group%2Frepo/resource_groups/application-rc-1986":
			payload, _ := io.ReadAll(r.Body)
			putBody = string(payload)
			putCalled = true
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := OperationService{SourceClient: GitLabClient{HTTPClient: server.Client()}}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}

	if err := service.Run(context.Background(), source, project, operationProcessModeOldestFirst); err != nil {
		t.Fatalf("run process mode operation: %v", err)
	}
	if !putCalled {
		t.Fatal("resource group PUT was not called")
	}
	if putBody != "process_mode=oldest_first" {
		t.Fatalf("body = %q", putBody)
	}
}

func TestOperationServiceReportsMissingResourceGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	service := OperationService{SourceClient: GitLabClient{HTTPClient: server.Client()}}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL}
	project := Project{ProviderID: "7"}

	err := service.Run(context.Background(), source, project, operationProcessModeOldestFirst)
	if err == nil || !strings.Contains(err.Error(), "application-rc-7") {
		t.Fatalf("err = %v, want missing resource group application-rc-7", err)
	}
}

func TestOperationServiceRejectsNonGitLabSources(t *testing.T) {
	service := OperationService{SourceClient: GitHubClient{}}

	err := service.Run(context.Background(), Source{Type: storage.SourceTypeGitHub}, Project{ProviderID: "owner/repo"}, operationShareCICache)
	if err == nil || !strings.Contains(err.Error(), "gitlab") {
		t.Fatalf("err = %v, want gitlab-only error", err)
	}
}

func TestGitLabClientUsesInjectedHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/api/v4/projects/group%2Frepo":
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		case "/api/v4/projects/group%2Frepo/repository/commits/main":
			_, _ = w.Write([]byte(`{"id":"abcdef1234567890"}`))
		case "/api/v4/projects/group%2Frepo/repository/files/package-lock.json/raw":
			if r.URL.Query().Get("ref") != "abcdef1234567890" {
				t.Fatalf("ref = %q", r.URL.Query().Get("ref"))
			}
			_, _ = w.Write([]byte(`{"lockfileVersion":3}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := GitLabClient{HTTPClient: server.Client()}
	source := Source{Type: storage.SourceTypeGitLab, URL: server.URL, PATToken: "secret"}
	project := Project{ProviderID: "group/repo"}
	commit, err := client.ResolveHead(context.Background(), source, project)
	if err != nil {
		t.Fatalf("resolve head: %v", err)
	}
	if commit.ShortSHA != "abcdef12" {
		t.Fatalf("short sha = %q, want abcdef12", commit.ShortSHA)
	}

	content, err := client.FetchFile(context.Background(), source, project, commit.SHA, "package-lock.json")
	if err != nil {
		t.Fatalf("fetch file: %v", err)
	}
	if string(content) != `{"lockfileVersion":3}` {
		t.Fatalf("content = %q", string(content))
	}
}

type fakeSourceClient struct {
	files                  map[string][]byte
	fetches                map[string]int
	divergence             BranchDivergence
	schedules              []PipelineSchedule
	ciSettings             CISettings
	numericProjectID       int64
	resourceGroupModes     map[string]string
	requestedResourceGroup string
}

func (c *fakeSourceClient) ResolveHead(context.Context, Source, Project) (Commit, error) {
	return Commit{SHA: "abcdef1234567890", ShortSHA: "abcdef12"}, nil
}

func (c *fakeSourceClient) FetchFile(_ context.Context, _ Source, _ Project, _ string, path string) ([]byte, error) {
	if c.fetches == nil {
		c.fetches = map[string]int{}
	}
	c.fetches[path]++

	content, ok := c.files[path]
	if !ok {
		return nil, ErrFileNotFound
	}

	return content, nil
}

func (c *fakeSourceClient) HasBranch(_ context.Context, _ Source, _ Project, branch string) (bool, error) {
	return branch == "main" || branch == "master" || branch == "dev", nil
}

func (c *fakeSourceClient) CompareBranches(context.Context, Source, Project, string, string) (BranchDivergence, error) {
	return c.divergence, nil
}

func (c *fakeSourceClient) DefaultBranch(_ context.Context, _ Source, _ Project) (string, error) {
	return "dev", nil
}

func (c *fakeSourceClient) ProtectedBranches(_ context.Context, _ Source, _ Project) ([]ProtectedBranch, error) {
	return nil, nil
}

func (c *fakeSourceClient) Tags(_ context.Context, _ Source, _ Project) ([]Tag, error) {
	return nil, nil
}

func (c *fakeSourceClient) PipelineSchedules(_ context.Context, _ Source, _ Project) ([]PipelineSchedule, error) {
	return c.schedules, nil
}

func (c *fakeSourceClient) CISettings(_ context.Context, _ Source, _ Project) (CISettings, error) {
	return c.ciSettings, nil
}

func (c *fakeSourceClient) NumericProjectID(_ context.Context, _ Source, _ Project) (int64, error) {
	if c.numericProjectID == 0 {
		return 1, nil
	}
	return c.numericProjectID, nil
}

func (c *fakeSourceClient) ResourceGroupProcessMode(_ context.Context, _ Source, _ Project, resourceGroup string) (string, error) {
	c.requestedResourceGroup = resourceGroup
	processMode, found := c.resourceGroupModes[resourceGroup]
	if !found {
		return "", ErrFileNotFound
	}
	return processMode, nil
}
