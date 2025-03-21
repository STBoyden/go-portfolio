package utils

import "strings"

func Cn(classes ...string) string {
	return strings.Join(classes, " ")
}
