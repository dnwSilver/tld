package projectsync

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	exportDefaultPattern = regexp.MustCompile(`\bexport\s+default\s+`)
	moduleExportsPattern = regexp.MustCompile(`\bmodule\.exports\s*=\s*`)
)

func isNextConfigCheck(checkID string) bool {
	switch checkID {
	case checkReactStrictMode, checkDistDir, checkOutput, checkValidateRSCRequestHeaders, checkSassCharset:
		return true
	default:
		return false
	}
}

func nextConfigCheckPasses(content []byte, checkID string) bool {
	config, _ := findNextConfigObject(stripJavaScriptComments(string(content)))

	switch checkID {
	case checkReactStrictMode:
		value, found := objectProperty(config, "reactStrictMode")
		return !found || booleanLiteral(value, true)
	case checkDistDir:
		value, found := objectProperty(config, "distDir")
		return found && stringLiteral(value, "dist")
	case checkOutput:
		value, found := objectProperty(config, "output")
		return found && stringLiteral(value, "standalone")
	case checkValidateRSCRequestHeaders:
		value, found := objectProperty(config, "validateRSCRequestHeaders")
		if !found {
			value, found = nestedObjectProperty(config, "experimental", "validateRSCRequestHeaders")
		}
		return found && booleanLiteral(value, false)
	case checkSassCharset:
		value, found := nestedObjectProperty(config, "sassOptions", "charset")
		return found && booleanLiteral(value, false)
	default:
		return false
	}
}

func nestedObjectProperty(object string, objectKey string, propertyKey string) (string, bool) {
	value, found := objectProperty(object, objectKey)
	if !found {
		return "", false
	}
	nested, found := objectLiteral(value)
	if !found {
		return "", false
	}
	return objectProperty(nested, propertyKey)
}

func findNextConfigObject(source string) (string, bool) {
	if match := exportDefaultPattern.FindStringIndex(source); match != nil {
		start := skipSpace(source, match[1])
		if object, found := objectAt(source, start); found {
			return object, true
		}
		if name, _, found := readIdentifier(source, start); found {
			if object, found := variableObject(source, name); found {
				return object, true
			}
		}
		if nameIndex := strings.Index(source[start:], "nextConfig"); nameIndex >= 0 {
			if object, found := variableObject(source, "nextConfig"); found {
				return object, true
			}
		}
	}

	if match := moduleExportsPattern.FindStringIndex(source); match != nil {
		if object, found := objectAt(source, skipSpace(source, match[1])); found {
			return object, true
		}
	}

	return variableObject(source, "nextConfig")
}

func variableObject(source string, name string) (string, bool) {
	pattern := regexp.MustCompile(`\b(?:const|let|var)\s+` + regexp.QuoteMeta(name) + `\b`)
	match := pattern.FindStringIndex(source)
	if match == nil {
		return "", false
	}

	equals := strings.IndexByte(source[match[1]:], '=')
	if equals < 0 {
		return "", false
	}
	initializer := skipSpace(source, match[1]+equals+1)
	if object, found := objectAt(source, initializer); found {
		return object, true
	}

	brace := strings.IndexByte(source[initializer:], '{')
	if brace < 0 {
		return "", false
	}
	return objectAt(source, initializer+brace)
}

func objectLiteral(value string) (string, bool) {
	return objectAt(value, skipSpace(value, 0))
}

func objectAt(source string, start int) (string, bool) {
	if start >= len(source) || source[start] != '{' {
		return "", false
	}

	depth := 0
	for index := start; index < len(source); index++ {
		switch source[index] {
		case '\'', '"', '`':
			index = stringEnd(source, index)
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[start : index+1], true
			}
		}
	}
	return "", false
}

func objectProperty(object string, wanted string) (string, bool) {
	if len(object) < 2 || object[0] != '{' {
		return "", false
	}

	for index := 1; index < len(object)-1; {
		index = skipDelimiters(object, index)
		if index >= len(object)-1 {
			break
		}

		key, next, found := readPropertyKey(object, index)
		if !found {
			index = nextProperty(object, index)
			continue
		}
		next = skipSpace(object, next)
		if next >= len(object) || object[next] != ':' {
			index = nextProperty(object, next)
			continue
		}

		valueStart := skipSpace(object, next+1)
		valueEnd := propertyValueEnd(object, valueStart)
		if key == wanted {
			return strings.TrimSpace(object[valueStart:valueEnd]), true
		}
		index = valueEnd + 1
	}

	return "", false
}

func readPropertyKey(source string, start int) (string, int, bool) {
	if start >= len(source) {
		return "", start, false
	}
	if source[start] == '\'' || source[start] == '"' {
		end := stringEnd(source, start)
		if end >= len(source) {
			return "", len(source), false
		}
		return source[start+1 : end], end + 1, true
	}
	return readIdentifier(source, start)
}

func readIdentifier(source string, start int) (string, int, bool) {
	if start >= len(source) || !identifierStart(rune(source[start])) {
		return "", start, false
	}
	end := start + 1
	for end < len(source) && identifierPart(rune(source[end])) {
		end++
	}
	return source[start:end], end, true
}

func identifierStart(value rune) bool {
	return value == '_' || value == '$' || unicode.IsLetter(value)
}

func identifierPart(value rune) bool {
	return identifierStart(value) || unicode.IsDigit(value)
}

func propertyValueEnd(source string, start int) int {
	braces := 0
	brackets := 0
	parentheses := 0
	for index := start; index < len(source); index++ {
		switch source[index] {
		case '\'', '"', '`':
			index = stringEnd(source, index)
		case '{':
			braces++
		case '}':
			if braces == 0 && brackets == 0 && parentheses == 0 {
				return index
			}
			braces--
		case '[':
			brackets++
		case ']':
			brackets--
		case '(':
			parentheses++
		case ')':
			parentheses--
		case ',':
			if braces == 0 && brackets == 0 && parentheses == 0 {
				return index
			}
		}
	}
	return len(source)
}

func nextProperty(source string, start int) int {
	end := propertyValueEnd(source, start)
	if end < len(source) {
		return end + 1
	}
	return end
}

func skipSpace(source string, start int) int {
	for start < len(source) && unicode.IsSpace(rune(source[start])) {
		start++
	}
	return start
}

func skipDelimiters(source string, start int) int {
	for start < len(source) && (unicode.IsSpace(rune(source[start])) || source[start] == ',') {
		start++
	}
	return start
}

func stringEnd(source string, start int) int {
	quote := source[start]
	for index := start + 1; index < len(source); index++ {
		if source[index] == '\\' {
			index++
			continue
		}
		if source[index] == quote {
			return index
		}
	}
	return len(source)
}

func booleanLiteral(value string, expected bool) bool {
	literal := "false"
	if expected {
		literal = "true"
	}
	fields := strings.Fields(strings.TrimSpace(value))
	return len(fields) > 0 && fields[0] == literal
}

func stringLiteral(value string, expected string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false
	}
	quote := value[0]
	return (quote == '\'' || quote == '"' || quote == '`') && value[len(value)-1] == quote && value[1:len(value)-1] == expected
}

func stripJavaScriptComments(source string) string {
	var result strings.Builder
	result.Grow(len(source))

	for index := 0; index < len(source); index++ {
		switch source[index] {
		case '\'', '"', '`':
			end := stringEnd(source, index)
			if end >= len(source) {
				result.WriteString(source[index:])
				return result.String()
			}
			result.WriteString(source[index : end+1])
			index = end
		case '/':
			if index+1 >= len(source) {
				result.WriteByte(source[index])
				continue
			}
			switch source[index+1] {
			case '/':
				result.WriteByte(' ')
				index += 2
				for index < len(source) && source[index] != '\n' {
					index++
				}
				if index < len(source) {
					result.WriteByte('\n')
				}
			case '*':
				result.WriteByte(' ')
				index += 2
				for index+1 < len(source) && !(source[index] == '*' && source[index+1] == '/') {
					if source[index] == '\n' {
						result.WriteByte('\n')
					}
					index++
				}
				index++
			default:
				result.WriteByte(source[index])
			}
		default:
			result.WriteByte(source[index])
		}
	}

	return result.String()
}
