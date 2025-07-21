package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tanay13/GlitchMesh/internal/models"
)

var Configs *models.Config

func Load(path string) (*models.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	var cfg models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode config JSON: %w", err)
	}

	Configs = &cfg

	return &cfg, nil
}
