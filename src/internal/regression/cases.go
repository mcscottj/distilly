// Package regression is the guard rail described in Milestone 2 of
// docs/roadmap.md: a suite of (original prompt, constraints that must
// survive optimization) pairs, checked before any AI-assisted rewriting
// (Milestone 3) is trusted to touch a prompt. It exists so that
// optimization never silently drops a constraint — e.g. collapsing
// "Always answer in JSON. Do not include markdown. Do not explain."
// down to "Return JSON."
package regression

// Case pairs a prompt fixture with the substrings that must still be
// present, verbatim, after running it through an optimizer.
type Case struct {
	Name string
	// PromptFile is relative to testdata/prompts.
	PromptFile string
	// MustSurvive lists substrings the optimized output must still
	// contain. Losing any one of these means the optimizer silently
	// dropped a constraint the original prompt required.
	MustSurvive []string
}

var Cases = []Case{
	{
		Name:       "exact duplicates: collapsing repeats keeps every distinct instruction",
		PromptFile: "exact_duplicates.txt",
		MustSurvive: []string{
			"Always respond in JSON format.",
			"Do not include markdown formatting in your response.",
			"Do not explain your reasoning, just answer.",
		},
	},
	{
		Name:       "near duplicates: cosmetic rewordings are not high-confidence, none may be silently dropped",
		PromptFile: "near_duplicates.txt",
		MustSurvive: []string{
			"Always respond in JSON format.",
			"Please always respond in JSON format.",
			"Do not explain your reasoning, just answer.",
			"Do not explain your reasoning; just answer",
		},
	},
	{
		Name:       "long history: system instruction and the actual question must survive untouched",
		PromptFile: "long_history.txt",
		MustSurvive: []string{
			"You are a helpful assistant.",
			"Given all that, how do I fix the networking timeout?",
		},
	},
	{
		Name:       "example.txt: textually-distinct phrasing of the same instruction is not high-confidence, none may be dropped",
		PromptFile: "example.txt",
		MustSurvive: []string{
			"Use markdown.",
			"Respond in markdown.",
			"Format as markdown.",
		},
	},
	{
		Name:       "redundant examples: only the exact whole-block duplicate collapses, distinct examples survive with their content intact",
		PromptFile: "redundant_examples.txt",
		MustSurvive: []string{
			"Q: What is the sum of 2 and 2?",
			`A: {"answer": 4}`,
			"Q: What's the sum of 2 and 2?",
			"Q: What is the capital of France?",
			`A: {"answer": "Paris"}`,
		},
	},
	{
		Name:       "structured data: JSON conversion is a format change, not just a length one, so it must not happen without approval",
		PromptFile: "structured_data.txt",
		MustSurvive: []string{
			"Name: John Smith",
			"Age: 30",
			"City: New York",
			"Occupation: Software Engineer",
		},
	},
}
