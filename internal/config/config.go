package config

import (
	"encoding/json"
	"fmt"
	"os"
)


func Load(path string) (*Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("could not open config file: %w", err)
    }
    defer file.Close()

    var cfg Config
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&cfg); err != nil {
        return nil, fmt.Errorf("could not decode config JSON: %w", err)
    }

    return &cfg, nil
}


