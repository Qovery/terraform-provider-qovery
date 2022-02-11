package descriptions

import (
	"fmt"
	"strings"
)

func NewStringEnumDescription(description string, enum []string, defaultValue *string) string {
	desc := fmt.Sprintf(
		"%s\n\t- Can be: `%s`.",
		description,
		strings.Join(enum, "`, `"),
	)

	if defaultValue != nil {
		desc += fmt.Sprintf(
			"\n\t- Default: `%s`.",
			*defaultValue,
		)
	}

	return desc
}
