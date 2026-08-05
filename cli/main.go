// Command spec is the Spec CLI's entry point: it hands the process's
// arguments and standard streams to the command tree cli/cmd builds, and
// exits with the status that tree decided on.
//
// This file holds nothing else. Every capability the binary offers —
// argument parsing, configuration discovery, output rendering, exit
// codes — lives in cli/cmd and the modules it composes, per the Command
// Surface component in specs/architecture/container-spec-cli.md.
package main

import (
	"os"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
