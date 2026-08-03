// Package tokenizer counts tokens using a real tiktoken-compatible BPE
// tokenizer. It uses the offline/embedded BPE loader so counting never
// makes a network call — consistent with distilly's local-first design.
package tokenizer

import (
	"strings"
	"sync"

	tiktoken "github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

// defaultEncoding is used when no model-specific encoding is requested.
// cl100k_base covers GPT-4/GPT-3.5 and is a reasonable general-purpose
// default for estimating prompt size.
const defaultEncoding = "cl100k_base"

var (
	once sync.Once
	enc  *tiktoken.Tiktoken
)

func init() {
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
}

func encoding() *tiktoken.Tiktoken {
	once.Do(func() {
		var err error
		enc, err = tiktoken.GetEncoding(defaultEncoding)
		if err != nil {
			// The offline loader embeds this encoding; a failure here means
			// the dependency is broken, not a runtime/input condition.
			panic("tokenizer: failed to load embedded " + defaultEncoding + " encoding: " + err.Error())
		}
	})
	return enc
}

// Count returns the number of tokens s encodes to under the default
// (cl100k_base) encoding.
func Count(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return len(encoding().Encode(s, nil, nil))
}

// CountForModel returns the number of tokens s encodes to under the
// encoding associated with model. ok is false if model isn't recognized,
// in which case the default encoding's count is returned as a fallback.
func CountForModel(model, s string) (count int, ok bool) {
	if strings.TrimSpace(s) == "" {
		return 0, true
	}
	tke, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return Count(s), false
	}
	return len(tke.Encode(s, nil, nil)), true
}
