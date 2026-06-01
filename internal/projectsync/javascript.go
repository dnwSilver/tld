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
	return []string{"package.json", "package-lock.json"}
}

func (JavaScriptStrategy) Parse(path string, content []byte) ([]Dependency, error) {
	if path != "package.json" {
		return []Dependency{}, nil
	}

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
	dependencies = appendManifestDependencies(dependencies, manifest.Dependencies, DependencyTypeRuntime, path)
	dependencies = appendManifestDependencies(dependencies, manifest.DevDependencies, DependencyTypeDev, path)
	dependencies = appendManifestDependencies(dependencies, manifest.PeerDependencies, DependencyTypePeer, path)
	dependencies = appendManifestDependencies(dependencies, manifest.OptionalDependencies, DependencyTypeOptional, path)
	dependencies = appendManifestDependencies(dependencies, manifest.Engines, DependencyTypeEngines, path)
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].DependencyType == dependencies[j].DependencyType {
			return dependencies[i].Name < dependencies[j].Name
		}
		return dependencies[i].DependencyType < dependencies[j].DependencyType
	})

	return dependencies, nil
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
