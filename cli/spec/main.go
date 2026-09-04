// Command spec is the Spec CLI's entry point: it hands the process's
// arguments and standard streams to the command tree cli/cmd builds, and
// exits with the status that tree decided on.
//
// This file holds nothing else. Every capability the binary offers —
// argument parsing, configuration discovery, output rendering, exit
// codes — lives in cli/cmd and the modules it composes, per the Command
// Surface component in specs/containers/spec-cli/spec-cli-container.md.
//
// It sits in cli/spec/ rather than at the module root because Go names
// an installed binary after the directory holding its main package: from
// here, `go install github.com/brettcallaghan/agent-plugins/cli/spec@…`
// installs a binary called `spec`, which is the tool's name. At the
// module root the same command would install one called `cli`. The Go
// module proxy is the registry AC-SPEC-009.6 is satisfied by, so the
// name that path produces is part of the deliverable, not cosmetic.
package main

import (
	"os"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
