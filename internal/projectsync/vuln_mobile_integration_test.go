package projectsync

import (
	"os"
	"testing"
)

func TestIOSVulnScanAvtocodWithTrivy(t *testing.T) {
	if os.Getenv("TLD_MOBILE_SCAN") != "1" {
		t.Skip("set TLD_MOBILE_SCAN=1 to run live mobile scan")
	}
	if _, err := os.Stat("/Users/kolosov.a/Repositories/avtocod/avtocod-ios/Gemfile.lock"); err != nil {
		t.Skip("avtocod-ios fixture is unavailable")
	}

	files := []File{
		{Path: "Podfile.lock", Content: mustReadFile(t, "/Users/kolosov.a/Repositories/avtocod/avtocod-ios/Podfile.lock")},
		{Path: "Gemfile.lock", Content: mustReadFile(t, "/Users/kolosov.a/Repositories/avtocod/avtocod-ios/Gemfile.lock")},
	}

	report, err := IOSVulnStrategy{}.Scan(files)
	if err != nil {
		t.Fatalf("scan ios: %v", err)
	}
	t.Logf("counts=%+v items=%d", report.Counts, len(report.Items))
	if report.Counts.High == 0 && report.Counts.Medium == 0 {
		t.Fatal("expected trivy gemfile vulnerabilities")
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}
