package sanitizer

import (
	"reflect"
)

// Sanitize dispatches v to the appropriate sanitisation handler based on its concrete type.
func (s *Sanitizer) Sanitize(v any) any {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case string:
		return s.SanitizeString(t)
	case map[string]any:
		return SanitizeMap(s, t)
	case map[string]string:
		return SanitizeMap(s, t)
	case []any:
		return SanitizeSlice(s, t)
	case []string:
		return SanitizeSlice(s, t)
	}

	return s.SanitizeReflect(v)
}

// SanitizeReflect handles types not covered by Sanitize's type switch via reflection.
func (s *Sanitizer) SanitizeReflect(v any) any {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Interface, reflect.Pointer:
		if rv.IsNil() {
			return nil
		}

		return s.Sanitize(rv.Elem().Interface())
	case reflect.Map:
		return s.sanitizeMapReflect(rv)

	case reflect.Slice, reflect.Array:
		return s.sanitizeSliceReflect(rv)

	case reflect.Struct:
		return s.sanitizeStructReflect(rv)

	case reflect.String:
		return s.SanitizeString(rv.String())
	}

	return v
}
