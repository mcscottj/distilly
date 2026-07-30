// Package tokenizer counts tokens in a way that approximates OpenAI's
// tiktoken. This is a placeholder word-based approximation; swap in a
// real tiktoken-compatible BPE tokenizer (e.g. github.com/pkoukk/tiktoken-go)
// once the lint pipeline shape is settled.
package tokenizer

import "strings"

// Count returns an approximate token count for s.
//
// Placeholder heuristic: ~4 characters per token, which is a commonly
// cited approximation for English text under GPT-family BPE tokenizers.
// Replace with a real tokenizer before trusting these numbers for
// billing-accuracy use cases.
func Count(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return (len(s) + 3) / 4
}
