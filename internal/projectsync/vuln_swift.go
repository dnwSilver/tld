package projectsync

import (
	"context"
	"errors"
	"fmt"
	"os"
)

type IOSVulnStrategy struct{}

func (IOSVulnStrategy) Files() []string {
	return []string{"Podfile.lock", "Podfile", "Gemfile.lock"}
}

func (IOSVulnStrategy) Scan(files []File) (VulnReport, error) {
	return (IOSVulnStrategy{}).ScanContext(context.Background(), files)
}

func (IOSVulnStrategy) ScanContext(ctx context.Context, files []File) (VulnReport, error) {
	if !hasIOSLockfile(files) {
		return VulnReport{}, errors.New("Podfile.lock or Gemfile.lock is required for ios vulnerability scan")
	}

	dir, err := os.MkdirTemp("", "tld-osv-ios-*")
	if err != nil {
		return VulnReport{}, fmt.Errorf("create ios scan temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := writeOsvScanFiles(dir, files); err != nil {
		return VulnReport{}, err
	}

	return runMobileVulnScanContext(ctx, dir, files, SwiftStrategy{})
}

func hasIOSLockfile(files []File) bool {
	for _, file := range files {
		if len(file.Content) == 0 {
			continue
		}
		if file.Path == "Podfile.lock" || file.Path == "Gemfile.lock" {
			return true
		}
	}
	return false
}
