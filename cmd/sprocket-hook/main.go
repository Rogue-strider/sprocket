// Command sprocket-hook is the PostToolUse hook binary. Claude Code pipes
// the hook payload to it on stdin; it runs a fast, language-appropriate
// check on the touched file and prints the result. See internal/hookrunner
// for the actual logic and the reasoning on why this never blocks or exits
// non-zero (PostToolUse hooks can't block — the tool already ran).
package main

import (
	"context"
	"os"

	"github.com/Rogue-strider/sprocket/internal/hookrunner"
)

func main() {
	// Always exit 0: a hook that fails internally must never break the
	// user's session. hookrunner.Run itself never returns a non-nil error
	// for this reason — this os.Exit(0) is just making that contract explicit
	// at the process boundary too.
	_ = hookrunner.Run(context.Background(), os.Stdin, os.Stdout, hookrunner.ExecRunner{})
	os.Exit(0)
}
