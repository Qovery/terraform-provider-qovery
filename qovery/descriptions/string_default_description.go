package descriptions

import (
	"fmt"
)

func NewStringDefaultDescription(description string, defaultValue string) string {
	return fmt.Sprintf(
		"%s\n\t- Default: `%s`.",
		description,
		defaultValue,
	)
}
