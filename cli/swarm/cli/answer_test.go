// Tests for the shape of a command's answer.
//
// The exit status selects the shape, so each case here fixes one status
// and asserts the document a reader holding that status would get.
package cli

import (
	"bytes"
	"testing"
)

func answerOf(t *testing.T, r Result) string {
	t.Helper()

	var out bytes.Buffer
	if err := renderAnswer(&out, r, false); err != nil {
		t.Fatalf("rendering the answer: %v", err)
	}
	return out.String()
}

var probeInvocation = []string{"bd", "show", "run-1", "--json"}

// A clear answer is the properties the command established and their
// values, in the order it established them, and nothing else.
func TestAClearAnswerCarriesTheEstablishedPropertiesInReportOrder(t *testing.T) {
	got := answerOf(t, Result{
		Outcome: OutcomeClear,
		Summary: "2 key(s) written",
		Facts: []Fact{
			Text("branch", "claude/branch-x", probeInvocation),
			Count("review_passes", 0, probeInvocation),
			Unestablished("trailers", "not written yet", probeInvocation),
		},
	})
	want := "branch: claude/branch-x\nreview_passes: 0\n"
	if got != want {
		t.Errorf("rendered %q, want %q", got, want)
	}
}

// A command that established nothing and is clear about it answers with
// an empty mapping rather than with nothing at all, so a consumer parses
// a document either way.
func TestAClearAnswerWithNoPropertiesIsAnEmptyMapping(t *testing.T) {
	if got := answerOf(t, Result{Outcome: OutcomeClear, Summary: "the manifest is empty"}); got != "{}\n" {
		t.Errorf("rendered %q, want %q", got, "{}\n")
	}
}

// Anything but a clear outcome renders a verdict per property: every
// property the command was asked about stays in the document, so the
// reader sees the whole set rather than a filtered one, and a failing
// property's issues name what would resolve it.
func TestAFailingAnswerCarriesAVerdictPerProperty(t *testing.T) {
	got := answerOf(t, Result{
		Outcome:    OutcomeAttention,
		Summary:    "findings_path: not written",
		Resolution: "write it with: swarm manifest set --run run-1 --key findings_path --value <value>",
		Facts: []Fact{
			Unestablished("findings_path", "the key is not written", probeInvocation),
			Text("trailers", "do re me", probeInvocation),
		},
	})
	want := "findings_path:\n" +
		"  ok: false\n" +
		"  value: null\n" +
		"  issues:\n" +
		"    - the key is not written\n" +
		"    - \"write it with: swarm manifest set --run run-1 --key findings_path --value <value>\"\n" +
		"trailers:\n" +
		"  ok: true\n" +
		"  value: do re me\n" +
		"  issues: []\n"
	if got != want {
		t.Errorf("rendered:\n%s\nwant:\n%s", got, want)
	}
}

// A refusal that named no property answers with `issues` alone. The
// document stays decodable because `issues` maps to a sequence where a
// property maps to a mapping.
func TestARefusalThatNamesNoPropertyCarriesIssuesAlone(t *testing.T) {
	got := answerOf(t, Result{
		Outcome: OutcomeMalformed,
		Summary: "swarm manifest get is scoped to one run; pass --run <run-epic-id>",
	})
	want := "issues:\n  - swarm manifest get is scoped to one run; pass --run <run-epic-id>\n"
	if got != want {
		t.Errorf("rendered %q, want %q", got, want)
	}
}

// A finding about the run rather than about any one property has nowhere
// per-property to go, so it is reported where a refusal reports one — and
// the properties are still all there, each with its verdict.
func TestAConditionNoPropertyCarriesIsReportedAsIssues(t *testing.T) {
	got := answerOf(t, Result{
		Outcome:    OutcomeAttention,
		Summary:    "no dispatch worktree stands; 1 bead(s) ready",
		Resolution: "dispatch a new batch, or check why none is running",
		Facts:      []Fact{Count("ready_front_depth", 1, probeInvocation)},
	})
	want := "issues:\n" +
		"  - no dispatch worktree stands; 1 bead(s) ready\n" +
		"  - dispatch a new batch, or check why none is running\n" +
		"ready_front_depth:\n" +
		"  ok: true\n" +
		"  value: 1\n" +
		"  issues: []\n"
	if got != want {
		t.Errorf("rendered:\n%s\nwant:\n%s", got, want)
	}
}
