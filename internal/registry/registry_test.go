package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestStableAcrossRegistryKinds(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		packageRef string
		path       string
		body       string
		want       string
	}{
		{name: "npm", kind: KindNPM, packageRef: "@scope/pkg", path: "/@scope%2Fpkg", body: `{"versions":{"1.9.0":{},"2.0.0-beta.1":{},"1.10.0":{}}}`, want: "1.10.0"},
		{name: "go", kind: KindGo, packageRef: "github.com/Acme/mod", path: "/github.com/%21acme/mod/@v/list", body: "v1.9.0\nv2.0.0-rc.1\nv1.10.0\n", want: "v1.10.0"},
		{name: "maven", kind: KindMaven, packageRef: "org.example:core", path: "/org/example/core/maven-metadata.xml", body: `<metadata><versioning><versions><version>1.9.0</version><version>2.0.0-SNAPSHOT</version><version>1.10.0</version></versions></versioning></metadata>`, want: "1.10.0"},
		{name: "rubygems", kind: KindRubyGems, packageRef: "rails", path: "/api/v1/versions/rails.json", body: `[{"number":"7.1.0","prerelease":false},{"number":"8.0.0.rc1","prerelease":true},{"number":"7.2.0","prerelease":false}]`, want: "7.2.0"},
		{name: "cocoapods", kind: KindCocoaPods, packageRef: "Firebase/Analytics", path: "/all_pods_versions_0_3_5.txt", body: "Firebase/11.9.0/12.0.0-beta.1/11.10.0\n", want: "11.10.0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.EscapedPath() != test.path {
					t.Fatalf("path = %q, want %q", r.URL.EscapedPath(), test.path)
				}
				if r.Header.Get("Authorization") != "Bearer secret" {
					t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
				}
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			client := Client{HTTPClient: server.Client()}
			got, err := client.LatestStable(context.Background(), Source{URL: server.URL, Token: "secret", Kind: test.kind}, test.packageRef)
			if err != nil {
				t.Fatalf("latest stable: %v", err)
			}
			if got != test.want {
				t.Fatalf("latest stable = %q, want %q", got, test.want)
			}
		})
	}
}

func TestLatestStableRejectsUnknownKind(t *testing.T) {
	_, err := (Client{}).LatestStable(context.Background(), Source{URL: "https://example.test", Kind: "unknown"}, "pkg")
	if err == nil || !strings.Contains(err.Error(), "unsupported registry kind") {
		t.Fatalf("error = %v", err)
	}
}
