package lint

import (
	"regexp"
	"strconv"
	"strings"
)

// keyValueRe matches a single "Key: value" line — a piece of structured
// data written as prose instead of a compact format.
var keyValueRe = regexp.MustCompile(`^\s*([A-Za-z][A-Za-z0-9 _-]{0,40}):\s+(.+?)\s*$`)

// MinStructuredRun is the minimum number of consecutive key/value lines
// before a run is flagged as structured data worth converting to JSON.
// One or two incidental "Label: value" lines are ordinary prose (e.g. a
// stray aside), not a data block, and shouldn't be flagged.
const MinStructuredRun = 3

// StructuredBlock is a run of consecutive "Key: value" lines detected in
// a prompt, worth converting to a single compact JSON object.
type StructuredBlock struct {
	Keys   []string
	Values []string
	// Raw holds the original lines, verbatim, for diffing and for
	// locating the block in the source prompt.
	Raw []string
}

// JSON renders b as a single-line JSON object, preserving key order.
// Values are always encoded as JSON strings rather than inferring
// numbers/booleans — simpler, always valid, and the token savings come
// from dropping the repeated line/label overhead, not from type
// inference.
func (b StructuredBlock) JSON() string {
	parts := make([]string, len(b.Keys))
	for i, k := range b.Keys {
		parts[i] = strconv.Quote(k) + ": " + strconv.Quote(b.Values[i])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// FindStructuredData scans lines for runs of >= MinStructuredRun
// consecutive "Key: value" lines.
//
// Callers should scope this to the System section only (see Run) —
// Examples ("Q: .../A: ...") and History ("User: .../Assistant: ...")
// both look like key/value runs but are turn-based dialogue, not data;
// converting them to JSON would mangle the prompt's meaning.
func FindStructuredData(lines []string) []StructuredBlock {
	var blocks []StructuredBlock
	var keys, values, raw []string

	flush := func() {
		if len(keys) >= MinStructuredRun {
			blocks = append(blocks, StructuredBlock{Keys: keys, Values: values, Raw: raw})
		}
		keys, values, raw = nil, nil, nil
	}

	for _, line := range lines {
		if m := keyValueRe.FindStringSubmatch(line); m != nil {
			keys = append(keys, m[1])
			values = append(values, m[2])
			raw = append(raw, line)
			continue
		}
		flush()
	}
	flush()

	return blocks
}
