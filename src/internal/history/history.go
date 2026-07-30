// Package history flags and (eventually) compresses long conversation
// history sections within a prompt — e.g. collapsing 42 turns down to
// a bullet-point summary that preserves key context.
package history

// Turn is a single message in a conversation history.
type Turn struct {
	Role    string
	Content string
}

// Flag reports whether a history of this many turns is worth flagging
// to the user as a compression candidate. The threshold is a starting
// point, not tuned against real data yet.
func Flag(turns []Turn) bool {
	const threshold = 10
	return len(turns) > threshold
}

// TODO(milestone 1): parse a raw prompt's history section into Turns.
// TODO(milestone 3): AI-assisted summarization that preserves key
// context (e.g. "User is debugging Docker. Using Go. Fixed auth issue.
// Current problem is networking.").
