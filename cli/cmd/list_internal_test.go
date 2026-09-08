// Whitebox tests for the `--type` flag's value type, which only the
// internal package can reach: [newListCmd] is unexported, and the
// property under test is what pflag makes of [typeFilterValue]'s
// declared type NAME rather than anything the command prints.
//
// Two claims, and they are the same claim from opposite sides. pflag
// routes a typed accessor by the string Type() returns, so declaring
// "stringSlice" for a type that is not one handed every value pflag's
// stringSlice converter — which panicked on the empty string an unset
// flag renders as, and silently read "[]" back as no values at all,
// losing exactly the empty occurrence this flag exists to preserve. A
// type name pflag does not know makes it refuse the read instead.
//
// The third assertion pins the cost of that choice at zero: the help
// output must not move, which is the property that rules out the
// remedies acting on String() or on the placeholder instead.
package cmd

import (
	"strings"
	"testing"
)

// TestTypeFlagTypedAccessorRefusesRatherThanPanics covers the flag
// unset and the flag set to the empty value. Before the fix the first
// panicked inside pflag's stringSlice converter (slice bounds out of
// range [:-1], from sval[1:len(sval)-1] over the empty string) and the
// second returned an empty slice with a nil error, reporting `--type ""`
// as indistinguishable from no filter at all.
func TestTypeFlagTypedAccessorRefusesRatherThanPanics(t *testing.T) {
	for _, tc := range []struct {
		name     string
		setEmpty bool
	}{
		{name: "flag unset"},
		{name: `flag set to ""`, setEmpty: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := "specs.toml"
			c := newListCmd(&ioStreams{}, &p)
			if tc.setEmpty {
				if err := c.Flags().Set("type", ""); err != nil {
					t.Fatalf(`Flags().Set("type", "") = %v, want nil`, err)
				}
			}

			// The read is wrapped so that a panic fails this test with a
			// message naming the flag, rather than crashing package cmd's
			// whole test binary and taking every other test with it.
			var (
				got []string
				err error
			)
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf(`Flags().GetStringSlice("type") panicked: %v; want a returned error`, r)
					}
				}()
				got, err = c.Flags().GetStringSlice("type")
			}()

			if err == nil {
				t.Errorf(`Flags().GetStringSlice("type") err = nil, want non-nil: pflag must refuse the typed read, not answer it — %q reads as "no filter"`, got)
			}
			if len(got) != 0 {
				t.Errorf(`Flags().GetStringSlice("type") = %q, want empty on a refused read`, got)
			}
		})
	}
}

// TestTypeFlagUsageLineUnchanged pins `specs list --help`'s flag line.
// The type name is unrecognised by pflag deliberately, and pflag renders
// an unrecognised name verbatim, so the placeholder stays `strings`. The
// absence of "(default" is the half that rules out returning "[]" from
// String(): pflag captures DefValue by calling String() at registration,
// and a non-zero default prints as `(default [])`.
func TestTypeFlagUsageLineUnchanged(t *testing.T) {
	p := "specs.toml"
	c := newListCmd(&ioStreams{}, &p)
	usages := c.Flags().FlagUsages()

	if !strings.Contains(usages, "--type strings") {
		t.Errorf("FlagUsages() = %q, want it to contain %q", usages, "--type strings")
	}
	if strings.Contains(usages, "(default") {
		t.Errorf("FlagUsages() = %q, want no %q — the --type flag has no default to print", usages, "(default")
	}
}
