// Tests for the invocations this implementation declares. They assert
// argv vectors rather than behaviour, because an argv is what the plan
// carries and what explain mode prints — and because the flags below were
// each chosen against a measurement that a future reader would otherwise
// have to re-take.
package beads

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// argv renders a step's declared invocation as a vector, which is only
// possible because every argument this implementation builds is a
// literal: it never takes one from an earlier step.
func argv(t *testing.T, s plan.Step) []string {
	t.Helper()

	out := []string{s.Executable}
	for _, a := range s.Args {
		if a.IsRef() {
			t.Fatalf("step %q takes an argument from step %q; this implementation builds only literals", s.Name, a.Step())
		}
		out = append(out, a.Literal())
	}
	return out
}

func joined(t *testing.T, s plan.Step) string {
	t.Helper()
	return strings.Join(argv(t, s), " ")
}

func TestTheInvocationsItDeclares(t *testing.T) {
	tr := Tracker{}

	cases := []struct {
		name string
		step plan.Step
		want string
	}{
		{"read one bead", tr.Bead("s", "run-1"),
			"bd show run-1 --json"},

		// --all and --limit 0 are both deliberate; see Children's doc.
		{"the beads directly under one", tr.Children("s", "run-1"),
			"bd list --parent run-1 --all --limit 0 --json"},

		{"the whole ready set", tr.Ready("s", "run-1", tracker.ReadyFilter{}),
			"bd ready --parent run-1 --json"},

		{"a filtered ready set", tr.Ready("s", "run-1", tracker.ReadyFilter{
			Assignee: "dev-1", ExcludeKinds: []string{"epic"}, ExcludeLabels: []string{"human"},
		}),
			"bd ready --parent run-1 --assignee dev-1 --exclude-type epic --exclude-label human --json"},

		// Status and assignee in one invocation: two would leave a bead
		// claimed by nobody in between.
		{"claim", tr.Claim("s", "run-1.4", "dev-1"),
			"bd update run-1.4 --status in_progress --assignee dev-1 --json"},

		{"finish", tr.SetState("s", "run-1.4", tracker.StateDone),
			"bd update run-1.4 --status closed --json"},

		{"block", tr.SetState("s", "run-1.4", tracker.StateBlocked),
			"bd update run-1.4 --status blocked --json"},

		{"label", tr.AddLabel("s", "run-1.4", "delta:TRK"),
			"bd update run-1.4 --add-label delta:TRK --json"},

		{"note", tr.Note("s", "run-1.4", "retro:proof measured"),
			"bd comment run-1.4 retro:proof measured --json"},

		// bd's own argument order reads backwards: `bd dep add A B` means
		// "A depends on B". The mapping happens here and nowhere else.
		{"blocked by", tr.BlockedBy("s", "run-1.4", "run-1.2"),
			"bd dep add run-1.4 run-1.2 --type blocks --json"},

		{"discovered from", tr.DiscoveredFrom("s", "run-1.20", "run-1.5"),
			"bd dep add run-1.20 run-1.5 --type discovered-from --json"},
	}

	for _, c := range cases {
		if got := joined(t, c.step); got != c.want {
			t.Errorf("%s:\n  got  %s\n  want %s", c.name, got, c.want)
		}
	}
}

// The comment body is one argument, whatever it contains. This is the
// argv-not-shell property arriving where it matters most: a finding's
// text is prose an agent wrote, and it routinely contains quotes,
// semicolons and dollar signs.
func TestANoteBodyIsOneArgumentWhateverItContains(t *testing.T) {
	const body = `retro:proof it's "measured"; $(echo hi) | tee x`

	got := argv(t, Tracker{}.Note("s", "run-1.4", body))
	if len(got) != 5 {
		t.Fatalf("the invocation has %d elements, want 5: %q", len(got), got)
	}
	if got[3] != body {
		t.Errorf("the body was altered:\n  got  %q\n  want %q", got[3], body)
	}
}

// Reads and writes are declared as such, because that declaration is the
// whole basis of dry-run mode: a write mis-declared as a read is one a
// dry run would perform.
func TestEachStepDeclaresItsEffectAndWhetherReRunningIsSafe(t *testing.T) {
	tr := Tracker{}

	reads := []plan.Step{
		tr.Bead("s", "run-1"),
		tr.Children("s", "run-1"),
		tr.Ready("s", "run-1", tracker.ReadyFilter{}),
	}
	for _, s := range reads {
		if s.Effect != plan.Read {
			t.Errorf("step %q is declared %s, want read", s.Description, s.Effect)
		}
		if !s.Repeatable {
			t.Errorf("read step %q is declared unsafe to re-run", s.Description)
		}
	}

	writes := []plan.Step{
		tr.Claim("s", "run-1.4", "dev-1"),
		tr.SetState("s", "run-1.4", tracker.StateDone),
		tr.AddLabel("s", "run-1.4", "x"),
		tr.Note("s", "run-1.4", "x"),
		tr.BlockedBy("s", "run-1.4", "run-1.2"),
		tr.DiscoveredFrom("s", "run-1.4", "run-1.2"),
	}
	for _, s := range writes {
		if s.Effect != plan.Write {
			t.Errorf("step %q is declared %s, want write", s.Description, s.Effect)
		}
	}

	// A note is the one write here that is not safe to re-run: running it
	// twice records the comment twice.
	if tr.Note("s", "run-1.4", "x").Repeatable {
		t.Error("a note is declared safe to re-run, but re-running it records the comment twice")
	}
	// The edge writes are, measured: re-adding an existing edge exits 0.
	if !tr.BlockedBy("s", "a", "b").Repeatable {
		t.Error("an edge write is declared unsafe to re-run, but re-adding an existing edge exits 0")
	}
}

// A state this tracker cannot write becomes a step that refuses at its
// decoder rather than a silently wrong write. The plan still renders, so
// explain shows what would have been attempted.
func TestAStateTheTrackerCannotWriteRefusesRatherThanGuessing(t *testing.T) {
	s := Tracker{}.SetState("s", "run-1.4", tracker.StateUnknown)

	if s.Executable == "" {
		t.Fatal("the refusing step is not renderable")
	}
	if s.Decode == nil {
		t.Fatal("the refusing step has no decoder to refuse in")
	}
	if _, err := s.Decode(swarmResultOK()); err == nil {
		t.Error("the step's decoder accepted a result for a state the tracker cannot write")
	}
}

// A caller may point the implementation at a particular bd and a
// particular working directory — bd resolves its database from the
// working directory, so a run reading a worktree's tracker sets it rather
// than changing its own process's directory.
func TestTheExecutableAndDirectoryAreConfigurable(t *testing.T) {
	tr := Tracker{Bd: "/opt/bin/bd", Dir: "/w/dev-1"}
	s := tr.Bead("s", "run-1")

	if s.Executable != "/opt/bin/bd" {
		t.Errorf("Executable = %q", s.Executable)
	}
	if s.Dir != "/w/dev-1" {
		t.Errorf("Dir = %q", s.Dir)
	}
}

// TestOnlyThisPackageNamesTheTool enforces "a component other than the
// Beads Tracker SHALL NOT invoke bd". Go has no mechanism for it, so this
// is the mechanism: no non-test source file under cli/swarm outside this
// package may contain the executable's name as a string literal.
//
// What it catches is a call site naming the tool directly, which is the
// way the rule actually gets broken. What it cannot catch is a name
// reaching an invocation through a variable — a textual guard has that
// limit, and the guard is still worth more than the sentence it replaces.
func TestOnlyThisPackageNamesTheTool(t *testing.T) {
	// The swarm tree, from this package's directory.
	root := filepath.Join("..", "..")
	thisPackage := filepath.Join("..", "..", "tracker", "beads")

	fset := token.NewFileSet()
	var scanned, hereToo int

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.HasPrefix(filepath.Clean(path), filepath.Clean(thisPackage)+string(filepath.Separator)) {
			hereToo++
			return nil
		}
		scanned++

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil || value != executable {
				return true
			}
			t.Errorf("%s names %q; only the Beads Tracker may invoke it", path, executable)
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking the swarm tree: %v", err)
	}

	// Guards against a vacuous pass: the walk has to have reached both
	// the rest of the tree and this package.
	if scanned == 0 {
		t.Fatalf("scanned no source outside this package under %s", root)
	}
	if hereToo == 0 {
		t.Fatalf("the walk did not reach this package, so its exemption is untested")
	}
}

// swarmResultOK is a successful, empty result — enough to drive a
// decoder that refuses before looking at anything.
func swarmResultOK() swarmexec.Result { return swarmexec.Result{} }
