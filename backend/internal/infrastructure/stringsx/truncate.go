package stringsx

// TruncateString shortens s to at most limit runes, appending "...[TRUNCATED]" if cut.
// Uses rune (Unicode code point) count — not bytes — so multi-byte characters
// (Cyrillic, Chinese, Arabic, etc.) are never split in the middle.
func TruncateString(val string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(val)
	if len(runes) <= limit {
		return val
	}
	return string(runes[:limit]) + "...[TRUNCATED]"
}
