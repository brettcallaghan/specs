// Command swarm is the Swarm CLI's entry point: it hands the process's
// arguments and standard streams to the command tree cli/swarm/cli
// builds, and exits with the status that tree decided on.
//
// This file holds nothing else. Argument parsing, plan rendering, output
// format and the exit vocabulary live in cli/swarm/cli and the packages
// it composes, per the Swarm Command Surface component in
// specs/architecture/container-swarm-cli.md.
//
// It sits in cli/swarm/ rather than at the module root because Go names
// an installed binary after the directory holding its main package: from
// here, `go install github.com/brettcallaghan/agent-plugins/cli/swarm@…`
// installs a binary called `swarm`, which is the tool's name. This is a
// second binary inside the existing cli module rather than a module of
// its own, per specs/adrs/ADR-023-go-for-the-swarm-cli.md, so it
// inherits that module's three-OS test matrix and six-target
// cross-compile instead of standing up a second set.
package main

import (
	"os"

	swarmcli "github.com/brettcallaghan/agent-plugins/cli/swarm/cli"
)

func main() {
	os.Exit(swarmcli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
