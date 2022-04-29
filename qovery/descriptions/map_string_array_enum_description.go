package descriptions

import (
	"fmt"
	"sort"
	"strings"
)

func NewMapStringArrayEnumDescription(description string, enum map[string][]string, defaultValue *string) string {
	desc := fmt.Sprintf("%s", description)

	keys := sortedMapKeys(enum)
	for _, key := range keys {
		values := enum[key]
		sort.Strings(values)
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

func sortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
