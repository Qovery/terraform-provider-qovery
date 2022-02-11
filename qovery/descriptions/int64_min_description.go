package descriptions

import (
	"fmt"
)

func NewInt64MinDescription(description string, min int64, defaultValue *int64) string {
	desc := fmt.Sprintf(
		"%s\n\t- Must be: `>= %d`.",
		description,
		min,
	)

	if defaultValue != nil {
		desc += fmt.Sprintf(
			"\n\t- Default: `%d`.",
			*defaultValue,
		)
	}

	return desc
}
