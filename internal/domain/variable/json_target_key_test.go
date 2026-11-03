//go:build unit && !integration
// +build unit,!integration

package variable

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJSONTargetKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, path string
		invalid    bool
	}{
		{"member", "password", false},
		{"nested", "db.credentials.password", false},
		{"array", "accounts.0.password", false},
		{"literal_dot", `database\.password`, false},
		{"literal_slash", `back\\slash.password`, false},
		{"literal_query_operators", `a*b?c|d#e@f`, false},
		{"literal_other_escape", `a\*b.password`, false},
		{"unicode_limit", strings.Repeat("🔑", 256), false},
		{"empty", "", true},
		{"blank", " ", true},
		{"leading_space", " password", true},
		{"trailing_space", "password ", true},
		{"empty_member", "a..b", true},
		{"leading_dot", ".password", true},
		{"trailing_dot", "password.", true},
		{"trailing_escape", `password\`, true},
		{"control", "a\nb", true},
		{"iso_control", "a\u0085b", true},
		{"over_limit", strings.Repeat("🔑", 257), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateJSONTargetKey(tc.path)
			if tc.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
