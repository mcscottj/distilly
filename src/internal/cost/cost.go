// Package cost estimates USD cost for a given token count against a
// table of known model prices. Prices are per 1,000 tokens and will
// drift out of date — treat this table as illustrative and update it
// before relying on it for real budgeting.
package cost

// Pricing holds $/1K token rates for a model.
type Pricing struct {
	Model          string
	InputPer1K     float64
	OutputPer1K    float64
}

// Table is a starter set of model prices. Extend as needed.
var Table = map[string]Pricing{
	"gpt-5": {
		Model:       "gpt-5",
		InputPer1K:  0.005,
		OutputPer1K: 0.015,
	},
}

// Estimate returns the estimated USD cost for inputTokens + outputTokens
// against the named model. ok is false if the model isn't in Table.
func Estimate(model string, inputTokens, outputTokens int) (usd float64, ok bool) {
	p, found := Table[model]
	if !found {
		return 0, false
	}
	usd = (float64(inputTokens)/1000)*p.InputPer1K + (float64(outputTokens)/1000)*p.OutputPer1K
	return usd, true
}
