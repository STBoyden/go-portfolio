package fs

import "embed"

//go:embed static
var StaticFS embed.FS

//go:embed migrations
var MigrationsFS embed.FS
