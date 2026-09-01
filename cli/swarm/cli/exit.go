package cli

import "fmt"

// Status is the process exit status a caller routes on. The four values
// are the contract FEAT-SWARM-CORE-001 fixes, and the same vocabulary the
// family's existing status script already uses.
type Status int

const (
	// StatusClear — the command answered, and its verdict is clear.
	StatusClear Status = 0
	// StatusAttention — the command answered, and what it found needs
	// attention. Its output names what would resolve the condition.
	StatusAttention Status = 1
	// StatusStop — the command found a condition the caller cannot repair
	// in place.
	StatusStop Status = 2
	// StatusMalformed — the invocation itself is malformed: a missing
	// scope, an unknown flag, a bad argument.
	StatusMalformed Status = 3
)

// Outcome is what a command concluded, in the run's own words. It is what
// a command reports; the status is derived from it rather than chosen
// alongside it, so the two cannot disagree.
type Outcome string

const (
	// OutcomeClear — it answered and its verdict is clear.
	OutcomeClear Outcome = "clear"
	// OutcomeAttention — it answered and what it found needs attention.
	OutcomeAttention Outcome = "attention"
	// OutcomeStop — it found a condition the caller cannot repair in
	// place.
	OutcomeStop Outcome = "stop"
	// OutcomeMalformed — its own invocation was malformed.
	OutcomeMalformed Outcome = "malformed"
)

// Outcomes returns the four outcomes, in status order. It is the
// enumeration a test asserts against, so "four and no others" is a fact
// about the code rather than a sentence about it.
func Outcomes() []Outcome {
	return []Outcome{OutcomeClear, OutcomeAttention, OutcomeStop, OutcomeMalformed}
}

// Status maps an outcome to the status a caller routes on.
//
// An unrecognised outcome maps to [StatusStop] rather than to
// [StatusClear]: a command whose outcome this package does not know is a
// condition the caller cannot repair in place, and defaulting to success
// would be the silent failure the whole container exists to remove.
func (o Outcome) Status() Status {
	switch o {
	case OutcomeClear:
		return StatusClear
	case OutcomeAttention:
		return StatusAttention
	case OutcomeStop:
		return StatusStop
	case OutcomeMalformed:
		return StatusMalformed
	default:
		return StatusStop
	}
}

// Refusal is a command declining to do what it was asked. Its outcome is
// chosen by what the caller can do about it, which is the distinction
// AC-SWARM-CORE-001.15 draws: bad arguments are the caller's to fix
// before re-invoking, a repairable state is theirs to fix in the
// repository, and an unrepairable one is not theirs at all.
type Refusal struct {
	// Outcome is what the refusal amounts to.
	Outcome Outcome
	// Reason is what was refused, in one line.
	Reason string
	// Resolution names what would resolve the condition. Required when
	// the outcome is [OutcomeAttention]; the surface refuses to render an
	// attention result without one.
	Resolution string
}

func (r *Refusal) Error() string { return r.Reason }

// RefuseArguments refuses because the arguments the command was given are
// not valid. The caller fixes the invocation and re-invokes.
func RefuseArguments(format string, a ...any) *Refusal {
	return &Refusal{Outcome: OutcomeMalformed, Reason: fmt.Sprintf(format, a...)}
}

// RefuseRepairable refuses because the state the command read is wrong
// and the caller can repair it in place. resolution names what would.
func RefuseRepairable(resolution, format string, a ...any) *Refusal {
	return &Refusal{Outcome: OutcomeAttention, Reason: fmt.Sprintf(format, a...), Resolution: resolution}
}

// RefuseUnrepairable refuses because the state the command read is wrong
// and the caller cannot repair it in place.
func RefuseUnrepairable(format string, a ...any) *Refusal {
	return &Refusal{Outcome: OutcomeStop, Reason: fmt.Sprintf(format, a...)}
}
