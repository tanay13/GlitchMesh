package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tanay13/GlitchMesh/internal/commands"
	"github.com/tanay13/GlitchMesh/internal/config"
)

func main() {
	_, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Error loading config %v", err)
	}

	config.ProxyLoad()

	cliArguments := os.Args[1:]

	if len(cliArguments) < 2 {
		PrintUsage()
		os.Exit(0)
	}

	cmdName := cliArguments[0]
	args := cliArguments[1:]

	for _, cmd := range commands.RegisteredCommands {
		if cmd.Name() == cmdName {

			if err := cmd.Execute(args); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Printf("Unknown command: %s\n", cmdName)
}

func PrintUsage() {}
