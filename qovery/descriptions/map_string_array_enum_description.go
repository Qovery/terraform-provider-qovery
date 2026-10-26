package descriptions

import (
	"fmt"
	"sort"
	"strings"
)

func NewMapStringArrayEnumDescription(description string, enum map[string][]string, defaultValue *string) string {
	var desc strings.Builder
	desc.WriteString(description)

	keys := sortedMapKeys(enum)
	for _, key := range keys {
		values := enum[key]
		sort.Strings(values)
		fmt.Fprintf(&desc,
			"\n\t- %s: `%s`.",
			key,
			strings.Join(values, "`, `"),
		)
	}

	if defaultValue != nil {
		fmt.Fprintf(&desc,
			"\n\t- Default: `%s`.",
			*defaultValue,
		)
	}

	return desc.String()
}

func sortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
