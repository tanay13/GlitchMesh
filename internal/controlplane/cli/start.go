package cli

import (
	"fmt"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/shared/constants"
	"github.com/tanay13/GlitchMesh/internal/dataplane/server"
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
	server.InitRouter()
	fmt.Println("Proxy server running on port 9000")
	http.HandleFunc("/", server.HomeHandler)
	http.HandleFunc("/metrics", server.MetricsHandler)
	http.HandleFunc("/redirect/", server.ProxyHandler)
	return http.ListenAndServe(":9000", nil)
}
