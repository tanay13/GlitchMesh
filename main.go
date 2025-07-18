package main

import (
	"os"

	"github.com/tanay13/GlitchMesh/internal/commands"
	"github.com/tanay13/GlitchMesh/internal/constants"
)

func main() {
	cliArguments := os.Args[1:]

	if len(cliArguments) < 2 {
		PrintUsage()
		os.Exit(0)
	}

	cmd := cliArguments[0]
	args := cliArguments[1:]

	switch cmd {
	case constants.CMD_START:
		commands.HandleStart(args)

	case constants.CMD_APPLY:
		commands.HandleApply(args)
	}
}

func PrintUsage() {}
