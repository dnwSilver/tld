package projectsync

import (
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const DependencyTypeGoModule = "require"

type GoStrategy struct{}

func (GoStrategy) Files() []string {
	return []string{"go.mod", "go.sum"}
}

func (GoStrategy) Parse(path string, content []byte) ([]Dependency, error) {
	if path != "go.mod" {
		return []Dependency{}, nil
	}

	return parseGoMod(content), nil
}

func parseGoMod(content []byte) []Dependency {
	lines := strings.Split(string(content), "\n")
	dependencies := make([]Dependency, 0)
	inRequireBlock := false
	for _, line := range lines {
		line = stripGoModComment(line)
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "require (" {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && trimmed == ")" {
			inRequireBlock = false
			continue
		}
		if inRequireBlock {
			dependencies = appendGoModDependency(dependencies, trimmed)
			continue
		}
		if strings.HasPrefix(trimmed, "require ") {
			dependencies = appendGoModDependency(dependencies, strings.TrimSpace(strings.TrimPrefix(trimmed, "require ")))
		}
	}

	return sortedDependencies(dependencies)
}

func stripGoModComment(line string) string {
	index := strings.Index(line, "//")
	if index == -1 {
		return line
	}

	return line[:index]
}

func appendGoModDependency(dependencies []Dependency, line string) []Dependency {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return dependencies
	}

	return append(dependencies, storage.ProjectDependency{
		Name:           strings.TrimSpace(parts[0]),
		Version:        strings.TrimSpace(parts[1]),
		DependencyType: DependencyTypeGoModule,
		SourceFile:     "go.mod",
	})
}
