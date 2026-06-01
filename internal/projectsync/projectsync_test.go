package projectsync

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
	projectEntity, err := store.Projects().Create(ctx, "owner/ios", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "iOS", "#EC9706")
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
	projectEntity, err := store.Projects().Create(ctx, "owner/go", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "Go App", "#EC9706")
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
	projectEntity, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, sourceEntity.ID, stack.ID, "󰏖", "TLD", "#EC9706")
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
	files   map[string][]byte
	fetches map[string]int
}

func (c *fakeSourceClient) ResolveHead(context.Context, Source, Project) (Commit, error) {
	return Commit{SHA: "abcdef1234567890", ShortSHA: "abcdef12"}, nil
}

func (c *fakeSourceClient) FetchFile(_ context.Context, _ Source, _ Project, _ string, path string) ([]byte, error) {
	if c.fetches == nil {
		c.fetches = map[string]int{}
	}
	c.fetches[path]++

	return c.files[path], nil
}

func (c *fakeSourceClient) HasBranch(_ context.Context, _ Source, _ Project, branch string) (bool, error) {
	return branch == "main" || branch == "master" || branch == "dev", nil
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
