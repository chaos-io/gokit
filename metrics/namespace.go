package metrics

import "strings"

const defaultNamespace = "service"

func normalizeNamespace(namespace string) string {
	var normalized strings.Builder
	normalized.Grow(len(namespace) + 1)

	for _, char := range namespace {
		switch {
		case char >= 'a' && char <= 'z',
			char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9',
			char == '_':
			normalized.WriteRune(char)
		default:
			normalized.WriteByte('_')
		}
	}

	if normalized.Len() == 0 {
		return defaultNamespace
	}
	value := normalized.String()
	if value[0] >= '0' && value[0] <= '9' {
		return "_" + value
	}
	return value
}
