package descriptions

import (
	"fmt"
	"strings"
)

func NewMapStringArrayEnumDescription(description string, enum map[string][]string, defaultValue *string) string {
	desc := fmt.Sprintf("%s", description)
	for key, values := range enum {
		desc += fmt.Sprintf(
			"\n\t- %s: `%s`.",
			key,
			strings.Join(values, "`, `"),
		)
	}

	if defaultValue != nil {
		desc += fmt.Sprintf(
			"\n\t- Default: `%s`.",
			*defaultValue,
		)
	}

	return desc
}
