package commands

import (
	"log"

	"github.com/tanay13/GlitchMesh/internal/utils"
)

func HandleApply(args []string) {
	if len(args) < 1 {
		log.Println("Not enough Arguments to apply")
		// PrintUsage()
	}

	filePath := args[0]

	utils.ParseYaml(filePath)
}
