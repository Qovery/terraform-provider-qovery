package descriptions

import (
	"fmt"
)

func NewInt64MinMaxDescription(description string, min int64, max int64, defaultValue *int64) string {
	desc := fmt.Sprintf(
		"%s\n\t- Must be: `>= %d` and `<= %d`.",
		description,
		min,
		max,
	)

	if defaultValue != nil {
		desc += fmt.Sprintf(
			"\n\t- Default: `%d`.",
			*defaultValue,
		)
	}

	return desc
}
