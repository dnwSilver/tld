package projectsync

import (
	"regexp"
	"strings"

	"github.com/dnwSilver/tld/internal/storage"
)

const (
	DependencyTypeGradleLibrary = "library"
	DependencyTypeGradlePlugin  = "plugin"
)

const kotlinVersionCatalog = "gradle/libs.versions.toml"

var kotlinBuildFiles = []string{"build.gradle.kts", "app/build.gradle.kts", "build.gradle", "app/build.gradle"}

var (
	tomlSectionPattern = regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	tomlVersionPattern = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=\s*"([^"]*)"\s*$`)
	tomlGroupPattern   = regexp.MustCompile(`group\s*=\s*"([^"]+)"`)
	tomlNamePattern    = regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
	tomlModulePattern  = regexp.MustCompile(`module\s*=\s*"([^"]+)"`)
	tomlIDPattern      = regexp.MustCompile(`id\s*=\s*"([^"]+)"`)
	tomlVersionRefPat  = regexp.MustCompile(`version\.ref\s*=\s*"([^"]+)"`)
	tomlInlineVerPat   = regexp.MustCompile(`version\s*=\s*"([^"]+)"`)

	gradleValPattern       = regexp.MustCompile(`(?m)^\s*(?:val|def)\s+([A-Za-z0-9_]+)\s*=\s*["']([^"']+)["']`)
	gradleConfigPattern    = regexp.MustCompile(`^"?([A-Za-z][A-Za-z0-9]*)"?[\s(]`)
	gradleCoordPattern     = regexp.MustCompile(`["']([A-Za-z0-9_.\-]+:[A-Za-z0-9_.\-]+(?::[^"']+)?)["']`)
	gradleDepBlockStartPat = regexp.MustCompile(`dependencies\s*\{`)
)

type KotlinStrategy struct{}

func (KotlinStrategy) Files() []string {
	return append([]string{kotlinVersionCatalog}, kotlinBuildFiles...)
}

func (KotlinStrategy) Parse(path string, content []byte) ([]Dependency, error) {
	switch {
	case path == kotlinVersionCatalog:
		return parseVersionCatalog(content), nil
	case strings.HasSuffix(path, ".gradle.kts"), strings.HasSuffix(path, ".gradle"):
		return parseGradleKts(path, content), nil
	default:
		return []Dependency{}, nil
	}
}

func parseVersionCatalog(content []byte) []Dependency {
	lines := strings.Split(string(content), "\n")
	versions := make(map[string]string)
	dependencies := make([]Dependency, 0)
	section := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if matches := tomlSectionPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
			section = matches[1]
			continue
		}

		switch section {
		case "versions":
			if matches := tomlVersionPattern.FindStringSubmatch(trimmed); len(matches) == 3 {
				versions[matches[1]] = matches[2]
			}
		case "libraries":
			if dependency, ok := parseCatalogLibrary(trimmed, versions); ok {
				dependencies = append(dependencies, dependency)
			}
		case "plugins":
			if dependency, ok := parseCatalogPlugin(trimmed, versions); ok {
				dependencies = append(dependencies, dependency)
			}
		}
	}

	return sortedDependencies(dependencies)
}

func parseCatalogLibrary(line string, versions map[string]string) (Dependency, bool) {
	name := ""
	if matches := tomlModulePattern.FindStringSubmatch(line); len(matches) == 2 {
		name = matches[1]
	} else {
		group := firstSubmatch(tomlGroupPattern, line)
		artifact := firstSubmatch(tomlNamePattern, line)
		if group == "" || artifact == "" {
			return Dependency{}, false
		}
		name = group + ":" + artifact
	}

	version := resolveCatalogVersion(line, versions)
	if version == "" {
		return Dependency{}, false
	}

	return storage.ProjectDependency{
		Name:           name,
		Version:        version,
		DependencyType: DependencyTypeGradleLibrary,
		SourceFile:     kotlinVersionCatalog,
	}, true
}

func parseCatalogPlugin(line string, versions map[string]string) (Dependency, bool) {
	id := firstSubmatch(tomlIDPattern, line)
	if id == "" {
		return Dependency{}, false
	}
	version := resolveCatalogVersion(line, versions)
	if version == "" {
		return Dependency{}, false
	}

	return storage.ProjectDependency{
		Name:           id,
		Version:        version,
		DependencyType: DependencyTypeGradlePlugin,
		SourceFile:     kotlinVersionCatalog,
	}, true
}

func resolveCatalogVersion(line string, versions map[string]string) string {
	if ref := firstSubmatch(tomlVersionRefPat, line); ref != "" {
		return versions[ref]
	}

	return firstSubmatch(tomlInlineVerPat, line)
}

func parseGradleKts(sourceFile string, content []byte) []Dependency {
	text := string(content)
	values := make(map[string]string)
	for _, match := range gradleValPattern.FindAllStringSubmatch(text, -1) {
		values[match[1]] = match[2]
	}

	dependencies := make([]Dependency, 0)
	for _, block := range gradleDependencyBlocks(text) {
		for _, line := range strings.Split(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}
			configMatch := gradleConfigPattern.FindStringSubmatch(trimmed)
			coordMatch := gradleCoordPattern.FindStringSubmatch(trimmed)
			if len(configMatch) != 2 || len(coordMatch) != 2 {
				continue
			}
			name, version, ok := splitGradleCoord(coordMatch[1], values)
			if !ok {
				continue
			}
			dependencies = append(dependencies, storage.ProjectDependency{
				Name:           name,
				Version:        version,
				DependencyType: configMatch[1],
				SourceFile:     sourceFile,
			})
		}
	}

	return sortedDependencies(dependencies)
}

func gradleDependencyBlocks(text string) []string {
	blocks := make([]string, 0)
	for _, loc := range gradleDepBlockStartPat.FindAllStringIndex(text, -1) {
		depth := 1
		index := loc[1]
		for index < len(text) && depth > 0 {
			switch text[index] {
			case '{':
				depth++
			case '}':
				depth--
			}
			index++
		}
		blocks = append(blocks, text[loc[1]:index-1])
	}

	return blocks
}

func splitGradleCoord(coord string, values map[string]string) (string, string, bool) {
	parts := strings.Split(coord, ":")
	if len(parts) < 3 {
		return "", "", false
	}
	version := resolveGradleVersion(parts[2], values)
	if version == "" {
		return "", "", false
	}

	return parts[0] + ":" + parts[1], version, true
}

func resolveGradleVersion(raw string, values map[string]string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "$") {
		return values[strings.Trim(raw[1:], "{}")]
	}

	return raw
}

func firstSubmatch(pattern *regexp.Regexp, line string) string {
	matches := pattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(matches[1])
}
