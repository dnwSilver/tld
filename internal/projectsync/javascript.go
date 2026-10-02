package projectsync

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	DependencyTypeRuntime  = "dependencies"
	DependencyTypeDev      = "devDependencies"
	DependencyTypePeer     = "peerDependencies"
	DependencyTypeOptional = "optionalDependencies"
	DependencyTypeEngines  = "engines"
	DependencyTypeNvmrc    = "nvmrc"
)

type StackStrategy interface {
	Files() []string
	Parse(path string, content []byte) ([]Dependency, error)
}

type JavaScriptStrategy struct{}

func ResolveStackStrategy(stackName string) (StackStrategy, error) {
	normalized := strings.ToLower(strings.TrimSpace(stackName))
	switch normalized {
	case "javascript", "js", "node", "nodejs", "node.js":
		return JavaScriptStrategy{}, nil
	case "go", "golang":
		return GoStrategy{}, nil
	case "swift", "ios":
		return SwiftStrategy{}, nil
	case "kotlin", "android":
		return KotlinStrategy{}, nil
	default:
		return nil, fmt.Errorf("unsupported stack %q", stackName)
	}
}

func (JavaScriptStrategy) Files() []string {
	return []string{"package.json", "package-lock.json", ".nvmrc"}
}

func (JavaScriptStrategy) Parse(path string, content []byte) ([]Dependency, error) {
	switch path {
	case "package.json":
		return parsePackageJSON(content)
	case ".nvmrc":
		return parseNvmrc(content)
	default:
		return []Dependency{}, nil
	}
}

func parsePackageJSON(content []byte) ([]Dependency, error) {
	var manifest struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
		Engines              map[string]string `json:"engines"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	dependencies := make([]Dependency, 0)
	dependencies = appendManifestDependencies(dependencies, manifest.Dependencies, DependencyTypeRuntime, "package.json")
	dependencies = appendManifestDependencies(dependencies, manifest.DevDependencies, DependencyTypeDev, "package.json")
	dependencies = appendManifestDependencies(dependencies, manifest.PeerDependencies, DependencyTypePeer, "package.json")
	dependencies = appendManifestDependencies(dependencies, manifest.OptionalDependencies, DependencyTypeOptional, "package.json")
	dependencies = appendManifestDependencies(dependencies, manifest.Engines, DependencyTypeEngines, "package.json")
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].DependencyType == dependencies[j].DependencyType {
			return dependencies[i].Name < dependencies[j].Name
		}
		return dependencies[i].DependencyType < dependencies[j].DependencyType
	})

	return dependencies, nil
}

func parseNvmrc(content []byte) ([]Dependency, error) {
	version, ok := readNvmrcVersion(content)
	if !ok {
		return []Dependency{}, nil
	}

	return []Dependency{{
		Name:           "node",
		Version:        version,
		DependencyType: DependencyTypeNvmrc,
		SourceFile:     ".nvmrc",
	}}, nil
}

func readNvmrcVersion(content []byte) (string, bool) {
	for _, line := range strings.Split(strings.ReplaceAll(string(content), "\r", ""), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		version := strings.TrimPrefix(line, "v")
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}

		return version, true
	}

	return "", false
}

func appendManifestDependencies(dependencies []Dependency, values map[string]string, dependencyType string, sourceFile string) []Dependency {
	for name, version := range values {
		dependencies = append(dependencies, storage.ProjectDependency{
			Name:           strings.TrimSpace(name),
			Version:        strings.TrimSpace(version),
			DependencyType: dependencyType,
			SourceFile:     sourceFile,
		})
	}

	return dependencies
}
