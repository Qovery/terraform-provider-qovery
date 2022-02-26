package qovery

import "strings"

func IsStatusError(state string) bool {
	return strings.HasSuffix(state, "_ERROR")
}
