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
		log.Fatalf("Error loading config: %v", err)
	}

	if err := config.ProxyLoad(); err != nil {
		log.Fatalf("Error loading proxy config: %v", err)
	}

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
	PrintUsage()
	os.Exit(1)
}

func PrintUsage() {
	fmt.Println("GlitchMesh 🧪 - A lightweight proxy for testing microservice resilience")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  glitchmesh <command> <subcommand>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start server    Start the proxy server")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  glitchmesh start server")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Edit config.json to set the path to your proxy YAML config file.")
	fmt.Println("  See README.md for YAML configuration format.")
}
