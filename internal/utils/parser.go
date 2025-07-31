package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/models"
	"gopkg.in/yaml.v3"
)

func ParseConfigYaml() (*models.Proxy, error) {

	filepath := config.Configs.Env.YAML_FILE_PATH

	if filepath == "" {
		log.Fatalf("Please set up the proxy yaml first!!")
	}

	t := models.Proxy{}

	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, fmt.Errorf("error while reading the file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &t)

	if err != nil {
		return nil, fmt.Errorf("error while parsing yaml: %v", err)
	}

	return &t, nil
}
