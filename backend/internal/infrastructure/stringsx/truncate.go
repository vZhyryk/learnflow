package stringsx

// TruncateString shortens s to at most limit runes (not bytes, so multi-byte chars
// aren't split), appending "...[TRUNCATED]" if cut.
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
