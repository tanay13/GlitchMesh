package commands

import (
	"fmt"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/router"
)

func HandleStart(args []string) {
	switch args[0] {
	case constants.SERVER:
		startServer()
	}
}

func startServer() {
	fmt.Println("Proxy server running on port 9000")
	http.HandleFunc("/", router.HomeHandler)
	http.HandleFunc("/redirect/", router.ProxyHandler)
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
