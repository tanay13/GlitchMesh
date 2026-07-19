package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
	"github.com/tanay13/GlitchMesh/internal/shared/utils"
)

var (
	mu               sync.RWMutex
	configs          *models.Config
	proxyConfig      *models.Proxy
	serviceOverrides = make(map[string]models.Fault)

	// configGeneration is incremented atomically on every successful hot-reload.
	configGeneration atomic.Int64

	reloadCallbacksMu sync.Mutex
	reloadCallbacks   []func(generation int64)
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

	if err := proxy.Validate(); err != nil {
		return fmt.Errorf("proxy config validation failed: %w", err)
	}

	gen := configGeneration.Add(1)

	mu.Lock()
	proxyConfig = proxy
	mu.Unlock()

	// Notify registered callbacks
	notifyReloadCallbacks(gen)

	return nil
}

// GetProxyConfig returns the current proxy config
func GetProxyConfig() *models.Proxy {
	mu.RLock()
	defer mu.RUnlock()
	return proxyConfig
}

// returns the JSON bootstrap config which contains the YAML path
func GetBootstrapConfig() *models.Config {
	mu.RLock()
	defer mu.RUnlock()
	return configs
}

// Incremented on every successful ProxyLoad().
func GetConfigGeneration() int64 {
	return configGeneration.Load()
}

// SetProxyConfigForTesting allows test code to inject a proxy config
// without loading from a file. Only use in tests.
func SetProxyConfigForTesting(proxy *models.Proxy) {
	mu.Lock()
	defer mu.Unlock()
	proxyConfig = proxy
}

// Override functions used by the admin APIs

// SetServiceOverride stores a runtime fault override for the named service.
// The override takes precedence over the YAML config until it is deleted.
func SetServiceOverride(serviceName string, fault models.Fault) error {
	if err := fault.Validate(); err != nil {
		return fmt.Errorf("invalid fault override for service %q: %w", serviceName, err)
	}
	mu.Lock()
	defer mu.Unlock()
	serviceOverrides[serviceName] = fault
	return nil
}

// DeleteServiceOverride removes the runtime override for a service,
// reverting it to whatever is in the YAML config.
func DeleteServiceOverride(serviceName string) {
	mu.Lock()
	defer mu.Unlock()
	delete(serviceOverrides, serviceName)
}

// GetEffectiveServiceConfig returns the merged service config for serviceName:
// override (if any) takes precedence over the YAML base config.
// Returns nil if the service is not found in either layer.
func GetEffectiveServiceConfig(serviceName string) *models.ServiceConfig {
	mu.RLock()
	defer mu.RUnlock()

	// Check the override layer first.
	if fault, ok := serviceOverrides[serviceName]; ok {
		// Find the base service to get its URL, then apply the overridden fault.
		if proxyConfig != nil {
			for _, svc := range proxyConfig.Service {
				if svc.Name == serviceName {
					merged := models.ServiceConfig{
						Name:  svc.Name,
						Url:   svc.Url,
						Fault: fault,
					}
					return &merged
				}
			}
		}
		// Service exists only in overrides (not in YAML) — not found.
		return nil
	}

	// Fall through to the base YAML config.
	if proxyConfig == nil {
		return nil
	}
	for _, svc := range proxyConfig.Service {
		if svc.Name == serviceName {
			s := svc // copy to avoid returning pointer into slice
			return &s
		}
	}
	return nil
}

// GetServiceOverrides returns a copy of the current override map.
// Used by the admin API's diff endpoint.
func GetServiceOverrides() map[string]models.Fault {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]models.Fault, len(serviceOverrides))
	for k, v := range serviceOverrides {
		result[k] = v
	}
	return result
}

// RegisterReloadCallback registers a function to be called each time the proxy
// config is successfully reloaded
func RegisterReloadCallback(fn func(generation int64)) {
	reloadCallbacksMu.Lock()
	defer reloadCallbacksMu.Unlock()
	reloadCallbacks = append(reloadCallbacks, fn)
}

func notifyReloadCallbacks(gen int64) {
	reloadCallbacksMu.Lock()
	defer reloadCallbacksMu.Unlock()
	for _, fn := range reloadCallbacks {
		fn(gen)
	}
}

// watcher function to class proxyLoad() when config file is changed/recreated
func StartWatcher(ctx context.Context, yamlPath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("could not create fsnotify watcher: %w", err)
	}

	if err := watcher.Add(yamlPath); err != nil {
		watcher.Close()
		return fmt.Errorf("could not watch %q: %w", yamlPath, err)
	}

	go func() {
		defer watcher.Close()
		log.Printf("[config] watching %q for changes (generation=%d)", yamlPath, configGeneration.Load())

		for {
			select {
			case <-ctx.Done():
				log.Printf("[config] watcher stopped (context cancelled)")
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					log.Printf("[config] detected change in %q, reloading...", event.Name)
					if err := ProxyLoad(); err != nil {
						log.Printf("[config] hot-reload failed: %v", err)
					} else {
						log.Printf("[config] hot-reload succeeded (generation=%d)", configGeneration.Load())
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[config] watcher error: %v", err)
			}
		}
	}()

	return nil
}
