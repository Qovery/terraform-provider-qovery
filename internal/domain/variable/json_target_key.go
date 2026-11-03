package variable

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateJSONTargetKey checks the member/index notation accepted by q-core without rewriting it.
func ValidateJSONTargetKey(value string) error {
	invalid := func(reason string) error { return fmt.Errorf("json_target_key %s", reason) }
	if value == "" {
		return invalid("must contain non-empty dot-separated members")
	}
	if strings.TrimSpace(value) != value {
		return invalid("must not start or end with whitespace")
	}
	if utf8.RuneCountInString(value) > 256 {
		return invalid("must not exceed 256 characters")
	}
	runes := []rune(value)
	memberEmpty := true
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if unicode.IsControl(c) {
			return invalid("must not contain control characters")
		}
		switch c {
		case '.':
			if memberEmpty {
				return invalid("must contain non-empty dot-separated members")
			}
			memberEmpty = true
		case '\\':
			if i+1 == len(runes) {
				return invalid("must not end with an unescaped backslash")
			}
			if runes[i+1] == '.' || runes[i+1] == '\\' {
				i++
			}
			memberEmpty = false
		default:
			memberEmpty = false
		}
	}
	if memberEmpty {
		return invalid("must contain non-empty dot-separated members")
	}
	return nil
}
