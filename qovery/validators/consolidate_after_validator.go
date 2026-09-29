package validators

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = consolidateAfterValidator{}

var consolidateAfterRegex = regexp.MustCompile(`^(\d+)(s|m|h)$`)

// consolidateAfterMaxSeconds is the largest Karpenter consolidate_after Qovery accepts: 24h.
const consolidateAfterMaxSeconds = 24 * 3600

var consolidateAfterUnitSeconds = map[string]int64{"s": 1, "m": 60, "h": 3600}

// consolidateAfterValidator validates a Karpenter node pool consolidate_after. Qovery accepts
// `<number><unit>` up to 24h, stores it in seconds and returns it in the largest whole unit (60m
// comes back as 1h, 120s as 2m), so only that form is accepted: any other spelling would come
// back different from the configuration and fail the apply with an inconsistent result.
type consolidateAfterValidator struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v consolidateAfterValidator) Description(_ context.Context) string {
	return "value must be <number><unit> with unit s, m or h, at most 24h, written in the largest whole unit (1h, not 60m)"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v consolidateAfterValidator) MarkdownDescription(_ context.Context) string {
	return "value must be `<number><unit>` with unit `s`, `m` or `h`, at most `24h`, written in the largest whole unit (`1h`, not `60m`)"
}

// ValidateString runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v consolidateAfterValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	match := consolidateAfterRegex.FindStringSubmatch(value)
	if match == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid consolidate_after",
			fmt.Sprintf("consolidate_after must be <number><unit> where unit is s (seconds), m (minutes) or h (hours), e.g. 30s, 10m or 1h, got: %q.", value),
		)
		return
	}

	unitSeconds := consolidateAfterUnitSeconds[match[2]]
	number, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || number > consolidateAfterMaxSeconds/unitSeconds {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid consolidate_after",
			fmt.Sprintf("consolidate_after must not exceed 24h, got: %q.", value),
		)
		return
	}

	if canonical := canonicalConsolidateAfter(number * unitSeconds); canonical != value {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid consolidate_after",
			fmt.Sprintf("Qovery stores consolidate_after %q as %q, in the largest whole unit. Write %q instead.", value, canonical, canonical),
		)
	}
}

// canonicalConsolidateAfter formats a number of seconds the way Qovery returns a consolidate_after.
func canonicalConsolidateAfter(seconds int64) string {
	switch {
	case seconds%3600 == 0:
		return fmt.Sprintf("%dh", seconds/3600)
	case seconds%60 == 0:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

func NewConsolidateAfterValidator() validator.String {
	return consolidateAfterValidator{}
}
