package descriptions

import (
	"fmt"
)

func NewInt64DefaultDescription(description string, defaultValue int64) string {
	return fmt.Sprintf(
		"%s\n\t- Default: `%d`.",
		description,
		defaultValue,
	)
}
