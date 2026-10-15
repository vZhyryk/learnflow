package sanitizer

import (
	"learnflow_backend/internal/infrastructure/stringsx"
	"regexp"
	"strings"
)

var opaquePathSegment = regexp.MustCompile(`^[A-Za-z0-9_\-.]{20,}$`)
var uuidPathSegment = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// SanitizePath redacts opaque-secret-looking path segments (e.g. a token in the URL)
// while leaving UUIDs untouched — a separate pass, since MaskInlineSecrets only sees "key=value" pairs.
func (s *Sanitizer) SanitizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if seg == "" || uuidPathSegment.MatchString(seg) {
			continue
		}
		if opaquePathSegment.MatchString(seg) {
			segments[i] = s.redactedValue
		}
	}
	return stringsx.TruncateString(strings.Join(segments, "/"), s.maxStringLen)
}

// MaskInlineSecrets scans val for known key=value patterns and replaces the values with the redacted marker.
func (s *Sanitizer) MaskInlineSecrets(val string) string {
	out := val

	for key := range s.sensitiveKeys {
		for _, variant := range s.KeyVariants(key) {
			markers := s.markersForVariant(key, variant)

			for _, marker := range markers {
				out = s.MaskAllWithMarker(out, marker)
			}
		}
	}

	return out
}

// MaskAllWithMarker replaces every value that immediately follows marker in val with the redacted marker.
func (s *Sanitizer) MaskAllWithMarker(val, marker string) string {
	if !strings.Contains(val, marker) {
		return val
	}

	var builder strings.Builder
	start := 0

	for {
		idx := strings.Index(val[start:], marker)
		if idx == -1 {
			builder.WriteString(val[start:])
			break
		}

		idx += start
		valueStart := idx + len(marker)

		builder.WriteString(val[start:valueStart])

		end := strings.IndexAny(val[valueStart:], " \n\r\t&?,;")
		builder.WriteString(s.redactedValue)

		if end == -1 {
			break
		}

		start = valueStart + end
	}

	return builder.String()
}

func (s *Sanitizer) markersForVariant(key, variant string) []string {
	switch normalizeSensitiveKey(key) {
	case "basic", "bearer":
		return []string{variant + " "}
	case "authorization", "proxyauthorization":
		return []string{variant + "="}
	default:
		return []string{
			variant + "=",
			variant + ": ",
			variant + ":",
			variant + " ",
		}
	}
}

// KeyVariants returns the lowercase, uppercase, title-case, and PascalCase forms of key.
func (s *Sanitizer) KeyVariants(key string) []string {
	seen := make(map[string]struct{})
	var out []string

	add := func(v string) {
		if v == "" {
			return
		}

		if _, ok := seen[v]; ok {
			return
		}

		seen[v] = struct{}{}
		out = append(out, v)
	}

	add(key)
	add(strings.ToUpper(key))
	if key != "" {
		add(strings.ToUpper(key[:1]) + strings.ToLower(key[1:]))
	}

	if strings.Contains(key, "_") || strings.Contains(key, "-") {
		add(stringsx.ToPascalFromSeparated(key))
	}

	return out
}
