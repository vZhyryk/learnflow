// Package ptr converts between wire-level DTOs ("" = not provided) and domain
// models (*string, nil = not provided / SQL NULL).
package ptr

// StringOrNil converts an unset (empty) wire-level string to nil for a
// nullable domain field. The inverse of StringOrEmpty.
func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringOrEmpty reads an optional domain field back into a plain string for
// contexts (e.g. event payloads) that don't carry the nil/"" distinction.
func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
