//go:build !ci

package fs

import "embed"

//go:embed .env
var EnvFile embed.FS
