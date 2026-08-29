// Package manifest is the run record's model: the declared key set, each
// key's type and the stage at which it becomes required, and the read and
// write path every other command reaches a run through — per the Manifest
// and Manifest Store components in
// specs/architecture/container-swarm-cli.md.
//
// It holds no knowledge of where the record lives. It asks a [Store] for
// the document and hands one back, which is what makes moving the record
// somewhere else a change to one small component rather than to the
// validation this container is mostly made of.
//
// The declared key set is not written here. It is read by reflection from
// the generated model, which is generated from
// schemas/swarm/manifest.strata — so a key's name, its type and the stage
// at which it becomes required all come from one declaration, and this
// package cannot hold a fourth opinion about any of them.
package manifest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// Key is one key of the declared set.
type Key struct {
	// Name is the key as it appears in the record.
	Name string
	// Stage is when it becomes required: the stage whose struct declares
	// it.
	Stage swarmmodel.RunStage
	// Type is the Go type of the key's value, with the optionality
	// stripped — the type a value has to unmarshal into.
	Type reflect.Type
}

// declared is the key set, built once from the generated model.
var declared = buildKeys()

func buildKeys() []Key {
	var keys []Key
	seen := map[string]bool{}

	for _, decl := range contract.StageDeclarations() {
		t := decl.Type
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name := wireName(field)
			if name == "" || seen[name] {
				// Inheritance flattens, so a later stage's struct repeats
				// every earlier key. The first stage carrying a key is the
				// stage it becomes required at.
				continue
			}
			seen[name] = true
			keys = append(keys, Key{Name: name, Stage: decl.Stage, Type: valueType(field.Type)})
		}
	}
	return keys
}

// wireName is the JSON key a generated field carries, empty when the
// field is not serialized.
func wireName(field reflect.StructField) string {
	tag, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}
	return name
}

// valueType strips the pointer the generator uses for an optional field,
// so the type reported is the type a value must unmarshal into rather
// than the type the struct happens to store it in.
func valueType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}

// Keys returns the declared key set, in stage order and then declaration
// order.
func Keys() []Key {
	out := make([]Key, len(declared))
	copy(out, declared)
	return out
}

// Lookup returns the declaration for a key, and whether it is declared at
// all.
func Lookup(name string) (Key, bool) {
	for _, k := range declared {
		if k.Name == name {
			return k, true
		}
	}
	return Key{}, false
}

// Document is the record as it was read: the raw object, so a key this
// version of the CLI does not know is neither lost nor reported as
// absent.
//
// The zero Document is an empty record, which is what a run epic with no
// manifest yet reads as.
type Document struct {
	raw map[string]json.RawMessage
}

// NewDocument returns a Document over the given object. The caller's map
// is copied, so a Document cannot change under whoever holds it.
func NewDocument(raw map[string]json.RawMessage) Document {
	copied := make(map[string]json.RawMessage, len(raw))
	for k, v := range raw {
		copied[k] = v
	}
	return Document{raw: copied}
}

// ParseDocument reads a Document out of a metadata object of the shape a
// tracker hands back.
func ParseDocument(metadata map[string]any) (Document, error) {
	raw := make(map[string]json.RawMessage, len(metadata))
	for k, v := range metadata {
		encoded, err := json.Marshal(v)
		if err != nil {
			return Document{}, fmt.Errorf("manifest: reading key %q: %w", k, err)
		}
		raw[k] = encoded
	}
	return Document{raw: raw}, nil
}

// Get returns a key's raw value and whether it is present.
//
// Present-and-absent are separate answers on purpose: a key that is
// absent is reported as absent rather than as an empty value of its
// declared type, because "not yet written" and "written as nothing" are
// different states of a run.
func (d Document) Get(key string) (json.RawMessage, bool) {
	v, ok := d.raw[key]
	return v, ok
}

// Names returns every key present, sorted.
func (d Document) Names() []string {
	out := make([]string, 0, len(d.raw))
	for k := range d.raw {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Raw returns a copy of the whole object.
func (d Document) Raw() map[string]json.RawMessage {
	return NewDocument(d.raw).raw
}

// Decode reads a key into v, reporting whether the key was present.
func (d Document) Decode(key string, v any) (bool, error) {
	value, ok := d.raw[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(value, v); err != nil {
		return true, fmt.Errorf("manifest: reading key %q: %w", key, err)
	}
	return true, nil
}

// Merged returns the document with changes applied at the top level, the
// way the record's own store merges. It does not mutate the receiver.
func (d Document) Merged(changes map[string]json.RawMessage) Document {
	merged := d.Raw()
	for k, v := range changes {
		merged[k] = v
	}
	return Document{raw: merged}
}

// Check reports every reason a set of changes may not be written: a key
// outside the declared set, or a value whose type is not its key's.
//
// Every reason, not the first — a caller that fixed one and re-ran to
// find the next would be doing the work this is for. And nothing is
// written when any of them holds: a multi-key write that failed on one
// key and stored the rest would leave the record in a state no component
// intended.
func Check(changes map[string]json.RawMessage) error {
	names := make([]string, 0, len(changes))
	for name := range changes {
		names = append(names, name)
	}
	sort.Strings(names)

	var refusals []string
	for _, name := range names {
		key, ok := Lookup(name)
		if !ok {
			refusals = append(refusals, fmt.Sprintf("%q is not a declared manifest key", name))
			continue
		}
		if err := json.Unmarshal(changes[name], newValue(key)); err != nil {
			refusals = append(refusals, fmt.Sprintf(
				"%q takes %s, and the value given is not one: %v", name, key.Type, err))
		}
	}
	if len(refusals) == 0 {
		return nil
	}
	return fmt.Errorf("manifest: refusing the write, and storing none of it — %s",
		strings.Join(refusals, "; "))
}

// newValue returns a fresh addressable zero value of a key's declared
// type, for unmarshalling a candidate value into.
func newValue(key Key) any { return reflect.New(key.Type).Interface() }

// Store reads and writes the manifest document, wherever it is kept.
//
// Conditional writing is in the interface rather than in the model
// because it is a property of the store: the refusal to clobber a
// concurrent change cannot be implemented above a store that does not
// offer it.
type Store interface {
	// Read declares the steps that fetch the record whole.
	// Answer: [Document].
	Read(prefix, run string) plan.Steps

	// Write declares the steps that apply changes to the record,
	// conditional on guard accepting the record as it stands at the
	// moment the write would happen. A guard that returns an error stops
	// the write.
	//
	// A predicate rather than a set of expected values, because more than
	// one condition needs the record as it stands: whether a key moved
	// since the caller read it, and whether the change breaks a rule that
	// can only be seen against what is already there. A store offering
	// only the first could not carry the second.
	//
	// The changes are applied as one merge, so a key the caller did not
	// set keeps whatever another writer gave it.
	// Answer: [Document], as written.
	Write(prefix, run string, changes map[string]json.RawMessage, guard Guard) plan.Steps
}

// Guard decides whether a write may proceed, given the record as it
// stands at the moment the write would happen. A nil Guard permits the
// write.
type Guard func(current Document) error

// Guards composes guards, running them in order and reporting the first
// refusal. A caller wanting every refusal at once composes the reports
// instead; this is for the write path, where the first reason to stop is
// enough to stop.
func Guards(guards ...Guard) Guard {
	return func(current Document) error {
		for _, g := range guards {
			if g == nil {
				continue
			}
			if err := g(current); err != nil {
				return err
			}
		}
		return nil
	}
}

// Unchanged returns the guard that refuses a write whose covered keys
// have moved since the caller read them.
func Unchanged(base Document, changes map[string]json.RawMessage) Guard {
	expect := Expect(base, changes)
	return func(current Document) error {
		if moved := Moved(expect, current, changes); len(moved) > 0 {
			return fmt.Errorf(
				"%s changed since it was read; re-read the manifest and try again",
				strings.Join(moved, ", "))
		}
		return nil
	}
}

// Absent returns the guard that refuses a write to a key already present.
// It is what --if-absent asks for: write this only if nobody has.
func Absent(changes map[string]json.RawMessage) Guard {
	return func(current Document) error {
		var present []string
		for name := range changes {
			if _, ok := current.Get(name); ok {
				present = append(present, name)
			}
		}
		sort.Strings(present)
		if len(present) > 0 {
			return fmt.Errorf("%s is already written, and this write asked to set it only if absent",
				strings.Join(present, ", "))
		}
		return nil
	}
}

// Rules returns the guard that refuses a write breaking a rule only
// visible against the record as it stands — a branch written twice, a
// counter that went backwards, a ledger entry removed.
func Rules(changes map[string]json.RawMessage) Guard {
	return func(current Document) error {
		if report := CheckWrite(current, changes); report.Failed() {
			return fmt.Errorf("%s", report.Error())
		}
		return nil
	}
}

// Expect returns the guard for a conditional write: the values the named
// keys held when the caller read the record.
func Expect(base Document, changes map[string]json.RawMessage) Document {
	guard := map[string]json.RawMessage{}
	for name := range changes {
		if value, ok := base.Get(name); ok {
			guard[name] = value
		}
	}
	return Document{raw: guard}
}

// Moved reports the keys whose value in current differs from what expect
// says it should be, over the keys the write covers.
//
// A key absent from expect is required to be absent from current: a key
// another writer added after this caller read is as much a concurrent
// change as one they altered.
func Moved(expect, current Document, changes map[string]json.RawMessage) []string {
	var moved []string
	for name := range changes {
		want, wanted := expect.Get(name)
		got, present := current.Get(name)
		switch {
		case wanted != present:
			moved = append(moved, name)
		case wanted && !sameJSON(want, got):
			moved = append(moved, name)
		}
	}
	sort.Strings(moved)
	return moved
}

// sameJSON compares two raw values by their decoded content, so a
// difference in whitespace or key order is not mistaken for a change.
func sameJSON(a, b json.RawMessage) bool {
	var left, right any
	if err := json.Unmarshal(a, &left); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &right); err != nil {
		return false
	}
	return reflect.DeepEqual(left, right)
}
