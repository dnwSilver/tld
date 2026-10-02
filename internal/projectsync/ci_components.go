package projectsync

import "strings"

const (
	checkCICD   = "ci/cd"
	checkNtfy   = "ntfy"
	checkDtrack = "dtrack"
	checkCremr  = "cremr"
)

func ciComponentVersions(content []byte) ProjectCheckVersions {
	versions := make(map[string][]string, 4)
	lines := strings.Split(strings.ReplaceAll(string(content), "\r", ""), "\n")

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if reference, found := yamlValue(trimmed, "component:"); found {
			separator := strings.LastIndexByte(reference, '@')
			if separator < 0 || separator == len(reference)-1 {
				continue
			}
			path := reference[:separator]
			version := reference[separator+1:]
			checkID := ciComponentCheckID(path)
			if checkID != "" {
				versions[checkID] = appendUniqueVersion(versions[checkID], strings.TrimSpace(version))
			}
			continue
		}

		if strings.Contains(trimmed, ciTemplatesMarker) {
			if version := legacyCITemplateRef(lines[index+1:]); version != "" {
				versions[checkCICD] = appendUniqueVersion(versions[checkCICD], version)
			}
		}
	}

	result := make(ProjectCheckVersions, len(versions))
	for checkID, values := range versions {
		result[checkID] = strings.Join(values, ",")
	}
	return result
}

func ciComponentCheckID(path string) string {
	path = strings.TrimSpace(path)
	switch {
	case strings.Contains(path, "/shared/ci-ntfy/"):
		return checkNtfy
	case strings.HasSuffix(path, "/shared/ci-security/dtrack-image"),
		strings.HasSuffix(path, "/shared/ci-security/dtrack-fs"):
		return checkDtrack
	case strings.Contains(path, "/shared/ci-mr/create-mr"):
		return checkCremr
	case strings.HasPrefix(lastPathPart(path), "build-"):
		return checkCICD
	default:
		return ""
	}
}

func lastPathPart(value string) string {
	if index := strings.LastIndexByte(value, '/'); index >= 0 {
		return value[index+1:]
	}
	return value
}

func yamlValue(line string, key string) (string, bool) {
	index := strings.Index(line, key)
	if index < 0 {
		return "", false
	}
	value := strings.TrimSpace(line[index+len(key):])
	if value == "" {
		return "", false
	}

	if value[0] == '\'' || value[0] == '"' {
		quote := value[0]
		if end := strings.IndexByte(value[1:], quote); end >= 0 {
			return value[1 : end+1], true
		}
	}
	if end := strings.IndexAny(value, " \t#"); end >= 0 {
		value = value[:end]
	}
	value = strings.TrimSpace(value)
	return value, value != ""
}

func legacyCITemplateRef(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if value, found := yamlValue(trimmed, "ref:"); found {
			return value
		}
		return ""
	}
	return ""
}

func appendUniqueVersion(versions []string, version string) []string {
	for _, existing := range versions {
		if existing == version {
			return versions
		}
	}
	return append(versions, version)
}
