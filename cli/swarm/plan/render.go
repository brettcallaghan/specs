package plan

import (
	"strings"
)

// refPlaceholder is how explain output renders an argument only an
// earlier step can supply. It names the step rather than a value,
// because learning the value would mean running the step explain mode
// exists not to run.
func refPlaceholder(step string) string { return "<from step: " + step + ">" }

// shellSafe is the set of characters a POSIX shell leaves alone. Anything
// outside it makes a word need quoting.
const shellSafe = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789" +
	"@%+=:,./-_"

// quote renders one argument as a single POSIX shell word.
//
// Explain output is a transcript a reader runs by hand, so each argument
// has to survive being pasted into a shell as the one argument the CLI
// would have passed. Single quotes, because they suppress every
// expansion; an embedded single quote is closed, escaped and reopened,
// which is the only way to put one inside a single-quoted word.
//
// The target is POSIX sh. On a Windows shell a quoted word carrying
// spaces reads differently, and a reader there adapts the transcript —
// which is a known cost of the transcript being text rather than an
// argv vector, not something quoting can fix.
func quote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(shellSafe, r)
	}) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// renderArgs renders a step's arguments for explain output, with each
// reference shown as the step that produces it.
func renderArgs(args []Arg) []string {
	out := make([]string, len(args))
	for i, a := range args {
		if a.IsRef() {
			out[i] = quote(refPlaceholder(a.step))
			continue
		}
		out[i] = quote(a.literal)
	}
	return out
}

// renderPlanned renders one step as the transcript line explain prints.
func renderPlanned(s Step) string {
	parts := append([]string{quote(s.Executable)}, renderArgs(s.Args)...)
	return strings.Join(parts, " ")
}

// renderResolved renders an invocation whose arguments are all known —
// what the executing interpreters print as they run each step.
func renderResolved(name string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quote(name))
	for _, a := range args {
		parts = append(parts, quote(a))
	}
	return strings.Join(parts, " ")
}

// Argv renders an argv vector — executable first, then each argument — as
// one POSIX sh command line, quoted so a reader can paste it.
//
// It is exported because every fact this CLI reports names the
// invocation that established it, and the surface rendering those facts
// must render an invocation exactly as the transcript does. Two
// renderings of the same argv would be two conventions to keep in step.
func Argv(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	return renderResolved(argv[0], argv[1:])
}

// annotation is the bracketed summary of a step's declarations, printed
// as a comment beside it. The three facts are the ones a reader cannot
// infer from the invocation and that change what the interpreters do.
func annotation(s Step) string {
	repeat := "not safe to re-run"
	if s.Repeatable {
		repeat = "safe to re-run"
	}
	onFail := "stop on failure"
	if s.OnFailure == Continue {
		onFail = "continue on failure"
	}
	return "[" + s.Effect.String() + "; " + onFail + "; " + repeat + "]"
}
