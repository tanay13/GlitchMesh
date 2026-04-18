package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/tanay13/GlitchMesh/internal/models"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

var (
	mu          sync.RWMutex
	configs     *models.Config
	proxyConfig *models.Proxy
)

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

	mu.Lock()
	configs = &cfg
	mu.Unlock()

	return &cfg, nil
}

func ProxyLoad() error {
	mu.RLock()
	yamlPath := configs.Env.YAML_FILE_PATH
	mu.RUnlock()

	proxy, err := utils.ParseConfigYaml(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to load proxy config: %w", err)
	}

	mu.Lock()
	proxyConfig = proxy
	mu.Unlock()

	return nil
}

func GetProxyConfig() *models.Proxy {
	mu.RLock()
	defer mu.RUnlock()
	return proxyConfig
}
