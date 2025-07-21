package utils

import (
	"fmt"
	"os"

	"github.com/tanay13/GlitchMesh/internal/models"
	"gopkg.in/yaml.v3"
)

func ParseYaml(filepath string) (*models.Proxy, error) {
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
