package projectsync

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScanOsvAPIFromGradleDependencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/querybatch":
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"vulns": [
							{"id": "GHSA-android-test"}
						]
					}
				]
			}`))
		case "/v1/vulns/GHSA-android-test":
			_, _ = w.Write([]byte(`{
				"id": "GHSA-android-test",
				"summary": "Android lib issue",
				"details": "Affected okhttp releases",
				"database_specific": {"severity": "HIGH"},
				"severity": [{"type":"CVSS_V3","score":"7.5"}]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	original := osvAPIBaseURL
	osvAPIBaseURL = server.URL
	defer func() {
		osvAPIBaseURL = original
	}()

	files := []File{{
		Path: "app/build.gradle.kts",
		Content: []byte(`
dependencies {
    implementation("com.squareup.okhttp3:okhttp:4.9.0")
}
`),
	}}

	report, err := scanOsvAPI(context.Background(), files, KotlinStrategy{})
	if err != nil {
		t.Fatalf("scan osv api: %v", err)
	}
	if !report.Scanned {
		t.Fatal("expected scanned report")
	}
	if len(report.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(report.Items))
	}
	if report.Items[0].Severity != VulnSeverityHigh {
		t.Fatalf("severity = %q, want high", report.Items[0].Severity)
	}
	if !strings.Contains(report.Items[0].Package, "okhttp") {
		t.Fatalf("package = %q", report.Items[0].Package)
	}
}

func TestOsvPackageRefCocoaPodsUsesSwiftURL(t *testing.T) {
	ecosystem, name, version, ok := osvPackageRef(Dependency{
		Name:           "Alamofire",
		Version:        "5.6.2",
		DependencyType: DependencyTypeCocoaPods,
	})
	if !ok {
		t.Fatal("expected cocoapods dependency to be queryable")
	}
	if ecosystem != "SwiftURL" || name != "Alamofire" || version != "5.6.2" {
		t.Fatalf("ref = %q %q %q", ecosystem, name, version)
	}
}

func TestScanOsvAPIFromPodfileLock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/querybatch":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"ecosystem":"SwiftURL"`) {
				t.Fatalf("expected SwiftURL ecosystem, got %s", body)
			}
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"vulns": [
							{"id": "GHSA-ios-test"}
						]
					}
				]
			}`))
		case "/v1/vulns/GHSA-ios-test":
			_, _ = w.Write([]byte(`{
				"id": "GHSA-ios-test",
				"summary": "Pod vulnerability",
				"database_specific": {"severity": "CRITICAL"},
				"severity": [{"type":"CVSS_V3","score":"9.1"}]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	original := osvAPIBaseURL
	osvAPIBaseURL = server.URL
	defer func() {
		osvAPIBaseURL = original
	}()

	files := []File{{
		Path: "Podfile.lock",
		Content: []byte(`PODS:
  - Alamofire (5.6.2)
DEPENDENCIES:
  - Alamofire
`),
	}}

	report, err := scanOsvAPI(context.Background(), files, SwiftStrategy{})
	if err != nil {
		t.Fatalf("scan osv api: %v", err)
	}
	if len(report.Items) != 1 || report.Items[0].Severity != VulnSeverityCritical {
		t.Fatalf("report = %#v", report)
	}
}

func TestExtractJSONPayload(t *testing.T) {
	output := []byte("Scanning dir /tmp/demo\n{\n  \"results\": []\n}")
	payload := extractJSONPayload(output)
	if !strings.HasPrefix(string(payload), "{") {
		t.Fatalf("payload = %q", payload)
	}
}
