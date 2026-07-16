// Package sanitizer redacts sensitive data from arbitrary Go values.
package sanitizer

import (
	"fmt"
	"learnflow_backend/internal/infrastructure/stringsx"
	"reflect"
	"strings"
)

// IsSensitiveKey reports whether key matches any known sensitive key after normalisation.
func (s *Sanitizer) IsSensitiveKey(key string) bool {
	key = normalizeSensitiveKey(key)
	_, ok := s.normalizedSensitiveKeys[key]
	return ok
}

func normalizeSensitiveKeys(keys map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(keys))

	for key := range keys {
		out[normalizeSensitiveKey(key)] = struct{}{}
	}

	return out
}

func normalizeSensitiveKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, " ", "")

	return key
}

// SanitizeString masks inline secrets in value and truncates it to the configured max length.
func (s *Sanitizer) SanitizeString(value string) string {
	return stringsx.TruncateString(s.MaskInlineSecrets(value), s.maxStringLen)
}

// pathLikeKeys are keys whose value is a URL path — redacted via SanitizePath instead of
// the generic key=value scan, since a path-embedded secret has no "key=value" marker.
var pathLikeKeys = map[string]struct{}{
	"path": {}, "urlpath": {}, "requestpath": {}, "requesturi": {},
}

func (s *Sanitizer) sanitizeMapValue(key string, value any) any {
	if s.IsSensitiveKey(key) {
		return s.redactedValue
	}

	if str, ok := value.(string); ok {
		if _, isPath := pathLikeKeys[normalizeSensitiveKey(key)]; isPath {
			return s.SanitizePath(str)
		}
	}

	return s.Sanitize(value)
}

func (s *Sanitizer) sanitizeMapReflect(rv reflect.Value) map[string]any {
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()

	for iter.Next() {
		key := fmt.Sprint(iter.Key().Interface())
		out[key] = s.sanitizeMapValue(key, iter.Value().Interface())
	}

	return out
}

func (s *Sanitizer) sanitizeSliceReflect(rv reflect.Value) []any {
	out := make([]any, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		out[i] = s.Sanitize(rv.Index(i).Interface())
	}

	return out
}

func (s *Sanitizer) sanitizeStructReflect(rv reflect.Value) map[string]any {
	out := make(map[string]any, rv.NumField())
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)
		if !fv.IsValid() || !fv.CanInterface() {
			continue
		}

		out[field.Name] = s.sanitizeMapValue(field.Name, fv.Interface())
	}

	return out
}

// SanitizeMap returns a copy of in with values for sensitive keys replaced by the redacted marker.
func SanitizeMap[T any](s *Sanitizer, in map[string]T) map[string]any {
	if in == nil {
		return nil
	}

	out := make(map[string]any, len(in))
	for k, val := range in {
		out[k] = s.sanitizeMapValue(k, val)
	}

	return out
}

// SanitizeSlice returns a copy of in with each element recursively sanitised.
func SanitizeSlice[T any](s *Sanitizer, in []T) []any {
	if in == nil {
		return nil
	}

	out := make([]any, len(in))
	for i, val := range in {
		out[i] = s.Sanitize(val)
	}

	return out
}
