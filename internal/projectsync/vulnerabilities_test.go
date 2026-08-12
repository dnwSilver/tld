package projectsync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestVulnerabilityReportsAreCachedPerMode(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	source := Source{Type: storage.SourceTypeGitHub}
	project := Project{ProviderID: "owner/repo"}
	prodReport := VulnReport{
		Scanned: true,
		Counts:  VulnCounts{High: 1},
		Items:   []Vulnerability{{Package: "runtime", Severity: VulnSeverityHigh}},
	}
	devReport := VulnReport{
		Scanned: true,
		Counts:  VulnCounts{High: 2},
		Items: []Vulnerability{
			{Package: "runtime", Severity: VulnSeverityHigh},
			{Package: "dev-only", Severity: VulnSeverityHigh},
		},
	}

	prodService := VulnScanService{Cache: store.Cache(), Mode: VulnScanModeProd}
	if err := prodService.cacheReport(ctx, source, project, prodReport); err != nil {
		t.Fatalf("cache prod report: %v", err)
	}
	devService := VulnScanService{Cache: store.Cache(), Mode: VulnScanModeDev}
	if err := devService.cacheReport(ctx, source, project, devReport); err != nil {
		t.Fatalf("cache dev report: %v", err)
	}

	loadedProd, err := prodService.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load prod report: %v", err)
	}
	if len(loadedProd.Items) != 1 || loadedProd.Items[0].Package != "runtime" {
		t.Fatalf("prod report = %#v", loadedProd)
	}

	loadedDev, err := devService.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load dev report: %v", err)
	}
	if len(loadedDev.Items) != 2 || loadedDev.Items[1].Package != "dev-only" {
		t.Fatalf("dev report = %#v", loadedDev)
	}
}
