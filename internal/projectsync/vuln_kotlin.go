package projectsync

import (
	"errors"
	"fmt"
	"os"
)

type AndroidVulnStrategy struct{}

func (AndroidVulnStrategy) Files() []string {
	return androidVulnFiles
}

func (AndroidVulnStrategy) Scan(files []File) (VulnReport, error) {
	if !hasGradleBuildFile(files) {
		return VulnReport{}, errors.New("gradle build file is required for android vulnerability scan")
	}

	dir, err := os.MkdirTemp("", "tld-osv-android-*")
	if err != nil {
		return VulnReport{}, fmt.Errorf("create android scan temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := writeOsvScanFiles(dir, files); err != nil {
		return VulnReport{}, err
	}

	return runMobileVulnScan(dir, files, KotlinStrategy{})
}
