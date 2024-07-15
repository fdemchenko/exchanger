package migrations

import "embed"

//go:embed rates/*.sql
var RatesMigrationsFS embed.FS
