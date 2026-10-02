package version

import (
	"strconv"
	"strings"
)

// CompareExact compares dotted numeric versions. Constraints, prereleases,
// aliases and ecosystem-specific formats are deliberately incomparable.
func CompareExact(left, right string) (comparison, leftMajor, rightMajor int, ok bool) {
	l, valid := parseExact(left)
	if !valid {
		return 0, 0, 0, false
	}
	r, valid := parseExact(right)
	if !valid {
		return 0, 0, 0, false
	}
	for index := 0; index < max(len(l), len(r)); index++ {
		lv, rv := 0, 0
		if index < len(l) {
			lv = l[index]
		}
		if index < len(r) {
			rv = r[index]
		}
		if lv < rv {
			return -1, l[0], r[0], true
		}
		if lv > rv {
			return 1, l[0], r[0], true
		}
	}
	return 0, l[0], r[0], true
}

func parseExact(raw string) ([]int, bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "v")
	if raw == "" {
		return nil, false
	}
	parts := strings.Split(raw, ".")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, false
			}
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		result = append(result, value)
	}
	return result, true
}
