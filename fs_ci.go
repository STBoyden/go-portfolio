//go:build ci

package fs

import (
	stdFS "io/fs"
)

//nolint:gochecknoglobals // This is a stub for CI purposes.
var EnvFile = stdFS.FS(nil)
