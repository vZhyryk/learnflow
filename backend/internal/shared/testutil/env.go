package testutil

import "testing"

// SetRequiredDBEnv sets the four DB_* vars db.BuildDSNFromEnv requires; t.Setenv restores
// each to its prior value when the test ends.
func SetRequiredDBEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PASSWORD", "testpass")
}
