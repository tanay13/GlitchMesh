package commands

import (
	"fmt"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/constants"
)

func HandleStart(args []string) {
	switch args[0] {
	case constants.SERVER:
		startServer()
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func startServer() {
	fmt.Println("Server running on http://localhost:9000")
	http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
