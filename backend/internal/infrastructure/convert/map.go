// Package convert provides type conversion helpers used across infrastructure layers.
package convert

import "encoding/json"

// ToMapStringAny converts any value to map[string]any.
// Returns (nil, false) if v is nil or conversion is not possible.
func ToMapStringAny(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m, ok := v.(map[string]string); ok {
		out := make(map[string]any, len(m))
		for k, val := range m {
			out[k] = val
		}
		return out, true
	}
	// JSON round-trip is the safest generic fallback: handles structs, custom map types,
	// and any other value that json.Marshal supports, without importing reflect directly.
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, false
	}
	return out, true
}
