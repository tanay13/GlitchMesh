package main

import (
	"log"
	"os"

	"github.com/tanay13/GlitchMesh/internal/commands"
	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/constants"
)

func main() {

	_, err := config.Load("config.json")

	if err != nil {
		log.Fatalf("Error loading config %v", err)
	}

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
	}
}

func PrintUsage() {}
