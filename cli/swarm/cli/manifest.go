package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// The `swarm manifest` commands: get, set, require and validate.
//
// Each is a [Spec] — a plan and an interpretation — so explain, dry-run
// and execute come from the Plan Runner rather than from four commands
// each having their own idea of what the modes mean.

// manifestCmd builds the `manifest` noun and its four verbs.
func manifestCmd(streams *ioStreams, store manifest.Store, runner swarmexec.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Read, write and check a run's manifest",
		Long: "The run manifest is the JSON object in a run epic's tracker metadata, " +
			"and the run's only durable record. These commands are how it is read and " +
			"written without destroying keys another component wrote.",
		SilenceUsage: true,
	}
	cmd.AddCommand(newSubcommand(manifestGet(store), streams, runner))
	cmd.AddCommand(newSubcommand(manifestSet(store), streams, runner))
	cmd.AddCommand(newSubcommand(manifestRequire(store), streams, runner))
	cmd.AddCommand(newSubcommand(manifestValidate(store), streams, runner))
	return cmd
}

// manifestGet reads the record, or one key of it.
//
// It is a lookup, so its answer is the value and not a report about the
// value: [Spec.Bare] makes `get --key branch` print the branch and
// nothing else, and --full-result is there for a caller that wants the
// summary and the provenance. The exit status says which of the three
// answers it gave — the value (0), a declared key nobody has written
// (1), or a key the manifest does not declare (3) — so the bare form
// loses no distinction the report carries.
func manifestGet(store manifest.Store) Spec {
	var keys []string

	return Spec{
		Use:       "get",
		Short:     "Read a run's manifest, or one key of it",
		RunScoped: true,
		Bare:      true,
		Flags: func(f *pflag.FlagSet) {
			f.StringArrayVar(&keys, "key", nil,
				"read only this key (repeatable); every key when omitted")
		},
		Plan: func(req Request) (plan.Plan, error) {
			if err := declared(keys); err != nil {
				return plan.Plan{}, err
			}
			return plan.Plan{Name: "read " + req.Run + "'s manifest", Steps: store.Read("read", req.Run)}, nil
		},
		Interpret: func(req Request, report plan.Report) (Result, error) {
			doc, from, err := document(store.Read("read", req.Run), report)
			if err != nil {
				return Result{}, err
			}

			wanted := keys
			if len(wanted) == 0 {
				wanted = declaredNames()
			}
			sort.Strings(wanted)

			result := Result{Outcome: OutcomeClear}
			var absent []string
			for _, name := range wanted {
				value, present := doc.Get(name)
				if !present {
					if len(keys) > 0 {
						absent = append(absent, name)
						result.Facts = append(result.Facts,
							Unestablished(name, "the key is not written", from))
					}
					continue
				}
				result.Facts = append(result.Facts, factFor(name, value, from))
			}

			switch {
			case len(absent) > 0:
				result.Outcome = OutcomeAttention
				result.Summary = fmt.Sprintf("%s: not written", strings.Join(absent, ", "))
				result.Resolution = fmt.Sprintf("write it with: swarm manifest set --run %s --key %s --value <value>",
					req.Run, absent[0])
				// No bare answer: there is no value to print, and the
				// status plus the reason on stderr is the whole answer.
				return result, nil
			case len(result.Facts) == 0:
				result.Summary = "the manifest is empty"
			default:
				result.Summary = fmt.Sprintf("%d key(s) written", len(result.Facts))
			}

			answer, err := lookupAnswer(doc, keys)
			if err != nil {
				return Result{}, err
			}
			result.Bare = answer
			return result, nil
		},
	}
}

// manifestSet writes keys, applying them as one merge.
func manifestSet(store manifest.Store) Spec {
	var (
		keys     []string
		values   []string
		ifAbsent bool
	)

	changes := func() (map[string]json.RawMessage, error) {
		if len(keys) == 0 {
			return nil, RefuseArguments("--key is required; pass --key <key> --value <value>")
		}
		if len(keys) != len(values) {
			return nil, RefuseArguments("%d --key and %d --value: each key takes one value, in order",
				len(keys), len(values))
		}
		out := map[string]json.RawMessage{}
		for i, name := range keys {
			if _, ok := out[name]; ok {
				return nil, RefuseArguments("%q is given twice in one invocation", name)
			}
			out[name] = json.RawMessage(values[i])
		}
		if err := manifest.Check(out); err != nil {
			return nil, RefuseArguments("%s", err)
		}
		return out, nil
	}

	return Spec{
		Use:       "set",
		Short:     "Write keys into a run's manifest, as one merge",
		RunScoped: true,
		Flags: func(f *pflag.FlagSet) {
			f.StringArrayVar(&keys, "key", nil, "the key to write (repeatable)")
			f.StringArrayVar(&values, "value", nil, "the value to write, as JSON (repeatable, paired with --key)")
			f.BoolVar(&ifAbsent, "if-absent", false, "write only if the key is not already written")
		},
		Plan: func(req Request) (plan.Plan, error) {
			applied, err := changes()
			if err != nil {
				return plan.Plan{}, err
			}
			guards := []manifest.Guard{manifest.Rules(applied)}
			if ifAbsent {
				guards = append(guards, manifest.Absent(applied))
			}
			return plan.Plan{
				Name:  fmt.Sprintf("write %s into %s's manifest", strings.Join(sorted(applied), ", "), req.Run),
				Steps: store.Write("set", req.Run, applied, manifest.Guards(guards...)),
			}, nil
		},
		Interpret: func(req Request, report plan.Report) (Result, error) {
			applied, err := changes()
			if err != nil {
				return Result{}, err
			}
			// The guard is what stops the plan, so its reason is the
			// answer — not the generic "the last step produced no value"
			// that reading the fragment's answer would give.
			if refusal := stopped(report); refusal != nil {
				return Result{}, refusal
			}

			steps := store.Write("set", req.Run, applied, nil)

			// A dry run performs the guard and refuses the write, so the
			// answer is what the guard READ — the record as it stands,
			// which is what "its report reflects the state the reads
			// returned" asks for. Reading the fragment's answer instead
			// would report "the write produced no value", which is true
			// and is not the question.
			if refused(report, steps) {
				doc, from, err := document(plan.Steps{steps[0]}, report)
				if err != nil {
					return Result{}, err
				}
				result := Result{
					Outcome: OutcomeClear,
					Summary: fmt.Sprintf("would write %s as one merge; nothing was written",
						strings.Join(sorted(applied), ", ")),
				}
				for _, name := range sorted(applied) {
					value, present := doc.Get(name)
					if !present {
						result.Facts = append(result.Facts,
							Unestablished(name, "the key is not written yet", from))
						continue
					}
					result.Facts = append(result.Facts, factFor(name, value, from))
				}
				return result, nil
			}

			doc, from, err := document(steps, report)
			if err != nil {
				return Result{}, err
			}

			result := Result{
				Outcome: OutcomeClear,
				Summary: fmt.Sprintf("wrote %s as one merge", strings.Join(sorted(applied), ", ")),
			}
			for _, name := range sorted(applied) {
				value, present := doc.Get(name)
				if !present {
					result.Facts = append(result.Facts,
						Unestablished(name, "the write reported no value for it", from))
					result.Outcome = OutcomeStop
					result.Summary = "the record does not carry what was written"
					continue
				}
				result.Facts = append(result.Facts, factFor(name, value, from))
			}
			return result, nil
		},
	}
}

// manifestRequire reports whether the named keys are written, and nothing
// about the rest of the record.
func manifestRequire(store manifest.Store) Spec {
	var keys []string

	return Spec{
		Use:       "require",
		Short:     "Report whether a run's manifest carries the named keys",
		RunScoped: true,
		Flags: func(f *pflag.FlagSet) {
			f.StringArrayVar(&keys, "key", nil, "a key the caller requires (repeatable)")
		},
		Plan: func(req Request) (plan.Plan, error) {
			if len(keys) == 0 {
				return plan.Plan{}, RefuseArguments("--key is required; pass one per key you require")
			}
			if err := declared(keys); err != nil {
				return plan.Plan{}, err
			}
			return plan.Plan{Name: "read " + req.Run + "'s manifest", Steps: store.Read("read", req.Run)}, nil
		},
		Interpret: func(req Request, report plan.Report) (Result, error) {
			doc, from, err := document(store.Read("read", req.Run), report)
			if err != nil {
				return Result{}, err
			}

			required := manifest.Require(doc, keys...)
			result := Result{Outcome: OutcomeClear}
			for _, name := range sortedCopy(keys) {
				if _, present := doc.Get(name); present {
					result.Facts = append(result.Facts, Flag(name, true, from))
					continue
				}
				result.Facts = append(result.Facts, Flag(name, false, from))
			}

			if required.Failed() {
				missing := namesOf(required.Failures())
				result.Outcome = OutcomeAttention
				result.Summary = fmt.Sprintf("%s: not written", strings.Join(missing, ", "))
				result.Resolution = fmt.Sprintf(
					"write each with: swarm manifest set --run %s --key <key> --value <value>", req.Run)
				return result, nil
			}
			result.Summary = fmt.Sprintf("every required key is written (%s)", strings.Join(sortedCopy(keys), ", "))
			return result, nil
		},
	}
}

// manifestValidate reports the whole record's health against every rule.
func manifestValidate(store manifest.Store) Spec {
	var stage string

	return Spec{
		Use:       "validate",
		Short:     "Check a run's manifest against every rule",
		RunScoped: true,
		Flags: func(f *pflag.FlagSet) {
			f.StringVar(&stage, "stage", "",
				"the stage to validate against; the stage the record has reached when omitted")
		},
		Plan: func(req Request) (plan.Plan, error) {
			if stage != "" && !knownStage(swarmmodel.RunStage(stage)) {
				return plan.Plan{}, RefuseArguments("%q is not a stage of a run; the stages are %s",
					stage, strings.Join(stageNames(), ", "))
			}
			return plan.Plan{Name: "read " + req.Run + "'s manifest", Steps: store.Read("read", req.Run)}, nil
		},
		Interpret: func(req Request, report plan.Report) (Result, error) {
			doc, from, err := document(store.Read("read", req.Run), report)
			if err != nil {
				return Result{}, err
			}

			at := swarmmodel.RunStage(stage)
			derived := stage == ""
			if derived {
				at = reached(doc)
			}

			validated := manifest.Validate(doc, at)
			result := Result{
				// The stage is a fact, not a footnote: a caller reading a
				// clean result has to know what it was clean against, and
				// when the stage was derived rather than given it is the
				// first thing they would ask.
				Facts: []Fact{Text("stage", string(at), from)},
			}
			if derived {
				result.Facts = append(result.Facts,
					Text("stage_source", "derived from the keys the record carries", from))
			}

			for _, f := range validated.Findings {
				name := strings.Join(f.Keys, ", ")
				if f.Severity == manifest.SeverityPending {
					result.Facts = append(result.Facts, Unestablished(name, f.Message, from))
					continue
				}
				result.Facts = append(result.Facts, Text(name+" ("+string(f.Rule)+")", f.Message, from))
			}

			if validated.Failed() {
				result.Outcome = OutcomeAttention
				result.Summary = fmt.Sprintf("%d rule(s) broken at the %s stage", len(validated.Failures()), at)
				result.Resolution = validated.Error()
				return result, nil
			}
			result.Outcome = OutcomeClear
			result.Summary = fmt.Sprintf("every rule holds at the %s stage", at)
			return result, nil
		},
	}
}

// --- shared helpers -------------------------------------------------

// stopped turns a plan the ordering guard stopped into the refusal its
// stopping step reported.
//
// A guard's whole job is to say why a write may not happen, so that
// reason has to reach the caller. Reading the fragment's answer instead
// would report "the last step produced no value", which is true and
// useless.
//
// The outcome is attention rather than stop: every condition a guard
// raises is one the caller can act on — re-read and retry after a
// concurrent change, or stop asking to change a key the run writes once.
func stopped(report plan.Report) *Refusal {
	if report.Stopped == "" {
		return nil
	}
	record, ok := report.Record(report.Stopped)
	if !ok || record.Err == nil {
		return nil
	}
	reason := record.Err.Error()
	return RefuseRepairable(reason, "%s", reason)
}

// refused reports whether a fragment's answering step was declined by
// the interpreter rather than run — which, for a write, is what a dry run
// does to it.
func refused(report plan.Report, steps plan.Steps) bool {
	record, ok := report.Record(steps.Answer())
	return ok && record.Status == plan.StepRefused
}

// document reads a fragment's answer as a manifest document, and returns
// the invocation that established it so every fact can name it.
func document(steps plan.Steps, report plan.Report) (manifest.Document, []string, error) {
	value, err := steps.ValueFrom(report)
	if err != nil {
		return manifest.Document{}, nil, RefuseUnrepairable("%s", err)
	}
	doc, ok := value.(manifest.Document)
	if !ok {
		return manifest.Document{}, nil, RefuseUnrepairable("the manifest read produced %T", value)
	}

	from := []string{}
	if record, ok := report.Record(steps.Answer()); ok && record.Invocation.Name != "" {
		from = append([]string{record.Invocation.Name}, record.Invocation.Args...)
	}
	// Every fact names the invocation that established it, so a fragment
	// whose record carried none is a defect rather than a fact with no
	// provenance.
	if len(from) == 0 {
		return manifest.Document{}, nil, RefuseUnrepairable("the manifest read recorded no invocation")
	}
	return doc, from, nil
}

// lookupAnswer is what `manifest get` answers when the caller wants the
// value rather than the report.
//
// Three shapes, because the question has three shapes. One key asks for a
// value and gets the value. No key asks for the record and gets the whole
// object. Several keys ask for part of the record and get an object of
// exactly those keys — not one value per line, which no reader could
// split back apart once a value carries a newline, and every manifest
// value may.
func lookupAnswer(doc manifest.Document, keys []string) (*Answer, error) {
	if len(keys) == 1 {
		value, present := doc.Get(keys[0])
		if !present {
			// Unreachable: an absent requested key took the attention
			// branch above. Refusing rather than printing nothing keeps a
			// future edit from turning "not written" into an empty answer
			// with a clear status.
			return nil, RefuseUnrepairable("the manifest read reported %q as present and then as absent", keys[0])
		}
		return valueAnswer(value)
	}

	wanted := doc.Raw()
	if len(keys) > 1 {
		wanted = map[string]json.RawMessage{}
		for _, name := range keys {
			if value, present := doc.Get(name); present {
				wanted[name] = value
			}
		}
	}
	return objectAnswer(wanted)
}

// valueAnswer renders one stored value.
//
// A stored string is printed as itself, without its JSON quoting: that is
// the whole point of the bare form, and it is what makes a value carrying
// whitespace, a newline or a leading dash arrive intact through command
// substitution. Everything else — a number, a boolean, a list, an object
// — is printed as compact JSON, which is both its own text rendering and
// a parseable one.
func valueAnswer(value json.RawMessage) (*Answer, error) {
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return nil, RefuseUnrepairable("the stored value is not JSON: %s", value)
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return nil, RefuseUnrepairable("the stored value cannot be re-encoded: %s", err)
	}
	if text, ok := decoded.(string); ok {
		return &Answer{Text: text, JSON: string(encoded)}, nil
	}
	return &Answer{Text: string(encoded), JSON: string(encoded)}, nil
}

// objectAnswer renders a set of keys as one JSON object, indented, with
// its keys in sorted order.
//
// The two forms are the same string: an object has no shell-friendlier
// rendering than JSON, so inventing one would give a reader two shapes to
// learn for one answer.
func objectAnswer(values map[string]json.RawMessage) (*Answer, error) {
	if values == nil {
		values = map[string]json.RawMessage{}
	}
	encoded, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return nil, RefuseUnrepairable("the manifest cannot be re-encoded: %s", err)
	}
	return &Answer{Text: string(encoded), JSON: string(encoded)}, nil
}

// factFor renders a manifest value as the fact kind matching its JSON.
//
// An object has no fact kind of its own, so it is reported as text
// carrying its JSON. That is the honest rendering: the contract's four
// kinds are what a consumer can branch on, and inventing a fifth here
// would put a shape in the envelope the schema does not declare.
func factFor(name string, value json.RawMessage, from []string) Fact {
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return Unestablished(name, "the stored value is not JSON", from)
	}

	switch v := decoded.(type) {
	case string:
		return Text(name, v, from)
	case bool:
		return Flag(name, v, from)
	case float64:
		if v == float64(int(v)) {
			return Count(name, int(v), from)
		}
	case []any:
		items := make([]string, 0, len(v))
		everyItemIsText := true
		for _, item := range v {
			text, ok := item.(string)
			if !ok {
				everyItemIsText = false
				break
			}
			items = append(items, text)
		}
		if everyItemIsText {
			return List(name, items, from)
		}
	case nil:
		return Unestablished(name, "the key is written as null", from)
	}
	return Text(name, compact(value), from)
}

func compact(value json.RawMessage) string {
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return string(value)
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return string(value)
	}
	return string(encoded)
}

// reached is the furthest stage every one of whose keys the record
// carries — what a caller means by "validate this run as it stands".
func reached(doc manifest.Document) swarmmodel.RunStage {
	at := swarmmodel.RunStagePreFlight
	for _, stage := range stagesInOrder() {
		complete := true
		for _, key := range manifest.Keys() {
			if key.Stage != stage {
				continue
			}
			if _, present := doc.Get(key.Name); !present {
				complete = false
				break
			}
		}
		if !complete {
			return at
		}
		at = stage
	}
	return at
}

func declared(keys []string) error {
	var unknown []string
	for _, name := range keys {
		if _, ok := manifest.Lookup(name); !ok {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		return RefuseArguments("%s: not a declared manifest key", strings.Join(unknown, ", "))
	}
	return nil
}

func declaredNames() []string {
	keys := manifest.Keys()
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k.Name)
	}
	return out
}

func sorted(changes map[string]json.RawMessage) []string {
	out := make([]string, 0, len(changes))
	for name := range changes {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func sortedCopy(names []string) []string {
	out := append([]string(nil), names...)
	sort.Strings(out)
	return out
}

func namesOf(findings []manifest.Finding) []string {
	var out []string
	for _, f := range findings {
		out = append(out, f.Keys...)
	}
	sort.Strings(out)
	return out
}

// stagesInOrder is the run's stages, in the order a run reaches them.
func stagesInOrder() []swarmmodel.RunStage { return contract.Stages() }

func knownStage(stage swarmmodel.RunStage) bool {
	for _, s := range stagesInOrder() {
		if s == stage {
			return true
		}
	}
	return false
}

func stageNames() []string {
	stages := stagesInOrder()
	out := make([]string, 0, len(stages))
	for _, s := range stages {
		out = append(out, string(s))
	}
	return out
}
