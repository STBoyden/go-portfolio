//go:build ci

package fs

import "github.com/halimath/fsmock"

//nolint:gochecknoglobals // This is a stub for CI purposes.
var EnvFile = fsmock.New(fsmock.NewDir("", fsmock.EmptyFile(".env")))
