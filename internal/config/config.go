package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tanay13/GlitchMesh/internal/models"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

var Configs *models.Config

var ProxyConfig *models.Proxy

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

func ProxyLoad() {
	ProxyConfig, _ = utils.ParseConfigYaml(Configs.Env.YAML_FILE_PATH)
}
