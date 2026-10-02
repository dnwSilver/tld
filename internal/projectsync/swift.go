package projectsync

import (
	"regexp"
	"sort"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	DependencyTypeBundler   = "bundler"
	DependencyTypeCocoaPods = "cocoapods"
)

var lockDependencyPattern = regexp.MustCompile(`^\s{2,}-\s+(.+?)\s+\(([^)]+)\)`)
var gemDependencyPattern = regexp.MustCompile(`^\s{4}([^\s(]+)\s+\(([^)]+)\)`)

type SwiftStrategy struct{}

func (SwiftStrategy) Files() []string {
	return []string{"Gemfile.lock", "Podfile.lock"}
}

func (SwiftStrategy) Parse(path string, content []byte) ([]Dependency, error) {
	switch path {
	case "Gemfile.lock":
		return parseGemfileLock(content), nil
	case "Podfile.lock":
		return parsePodfileLock(content), nil
	default:
		return []Dependency{}, nil
	}
}

func parseGemfileLock(content []byte) []Dependency {
	lines := strings.Split(string(content), "\n")
	dependencies := make([]Dependency, 0)
	inSpecs := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "specs:":
			inSpecs = true
			continue
		case "PLATFORMS", "DEPENDENCIES", "BUNDLED WITH":
			inSpecs = false
		}
		if !inSpecs {
			continue
		}

		matches := gemDependencyPattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		dependencies = append(dependencies, storage.ProjectDependency{
			Name:           strings.TrimSpace(matches[1]),
			Version:        strings.TrimSpace(matches[2]),
			DependencyType: DependencyTypeBundler,
			SourceFile:     "Gemfile.lock",
		})
	}

	return sortedDependencies(dependencies)
}

func parsePodfileLock(content []byte) []Dependency {
	lines := strings.Split(string(content), "\n")
	dependencies := make([]Dependency, 0)
	inPods := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "PODS:":
			inPods = true
			continue
		case "DEPENDENCIES:", "SPEC REPOS:", "EXTERNAL SOURCES:", "CHECKOUT OPTIONS:", "SPEC CHECKSUMS:", "PODFILE CHECKSUM:", "COCOAPODS:":
			inPods = false
		}
		if !inPods {
			continue
		}

		matches := lockDependencyPattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		dependencies = append(dependencies, storage.ProjectDependency{
			Name:           strings.TrimSpace(matches[1]),
			Version:        strings.TrimSpace(matches[2]),
			DependencyType: DependencyTypeCocoaPods,
			SourceFile:     "Podfile.lock",
		})
	}

	return sortedDependencies(dependencies)
}

func sortedDependencies(dependencies []Dependency) []Dependency {
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].DependencyType == dependencies[j].DependencyType {
			return dependencies[i].Name < dependencies[j].Name
		}

		return dependencies[i].DependencyType < dependencies[j].DependencyType
	})

	return dependencies
}
