package descriptions

import (
	"fmt"
)

func NewBoolDefaultDescription(description string, defaultValue bool) string {
	return fmt.Sprintf(
		"%s\n\t- Default: `%t`.",
		description,
		defaultValue,
	)
}
