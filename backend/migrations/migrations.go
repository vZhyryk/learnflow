package migrations

import "embed"

//go:embed *.sql

// FS embeds all SQL migration files for use by the migrate runner.
var FS embed.FS
