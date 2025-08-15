package commands

import (
	"fmt"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/router"
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
	fmt.Println("Proxy server running on port 9000")
	http.HandleFunc("/", router.HomeHandler)
	http.HandleFunc("/redirect/", router.ProxyHandler)
	return http.ListenAndServe(":9000", nil)
}
