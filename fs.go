package fs

import "embed"

//go:embed static
var StaticFS embed.FS

//go:embed .env
var EnvFile embed.FS
