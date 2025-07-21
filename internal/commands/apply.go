package commands

import (
	"log"
	"os"

	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func HandleApply(args []string) {
	if len(args) < 1 {
		log.Println("Not enough Arguments to apply")
		// PrintUsage()
	}

	filePath := args[0]
	os.Setenv(config.Configs.Env.YAML_FILE_PATH, filePath)
	utils.ParseYaml(filePath)
}
