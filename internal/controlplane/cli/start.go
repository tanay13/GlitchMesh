package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tanay13/GlitchMesh/internal/controlplane/adminapi"
	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/dataplane/server"
	"github.com/tanay13/GlitchMesh/internal/shared/constants"
)

type Start struct{}

func (c *Start) Name() string {
	return constants.CMD_START
}

func (c *Start) Execute(args []string) error {
	switch args[0] {
	case constants.SERVER:
		return c.startServer()
	default:
		return fmt.Errorf("unknown start type: %s", args[0])
	}
}

func (c *Start) startServer() error {
	// Context: cancelled on SIGTERM/SIGINT for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start fsnotify watcher so config changes are picked up without restart
	proxyConf := config.GetProxyConfig()
	if proxyConf == nil {
		return fmt.Errorf("proxy config not loaded — cannot start watcher")
	}

	// Retrieve YAML path from the bootstrap JSON config
	cfg := config.GetBootstrapConfig()
	if cfg != nil && cfg.Env.YAML_FILE_PATH != "" {
		if err := config.StartWatcher(ctx, cfg.Env.YAML_FILE_PATH); err != nil {
			log.Printf("[start] warning: could not start config watcher: %v", err)
		} else {
			log.Printf("[start] hot-reload watcher active on %q", cfg.Env.YAML_FILE_PATH)
		}
	}

	// Register log callback so reload events appear in the server log.
	config.RegisterReloadCallback(func(gen int64) {
		log.Printf("[config] config reloaded — generation=%d", gen)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.HomeHandler)
	mux.HandleFunc("/metrics", server.MetricsHandler)
	mux.HandleFunc("/redirect/", server.ProxyHandler)

	adminapi.RegisterRoutes(mux)

	server.InitRouter()

	srv := &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}

	// Run server in a goroutine so we can listen for shutdown signals
	errCh := make(chan error, 1)
	go func() {
		fmt.Println("Proxy server running on port 9000")
		fmt.Println("Admin API available at /admin/* (set ADMIN_TOKEN env var to protect it)")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Println("[start] shutting down gracefully...")
		if err := srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		log.Println("[start] server stopped")
		return nil
	}
}
