package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	cliArguments := os.Args[1:]

	if len(cliArguments) >= 2 && cliArguments[0] == "start" {
		switch cliArguments[1] {
		case "server":
			runServer()
		default:
			fmt.Println("Unknown command:", cliArguments[1])
		}
	} else {
		fmt.Println("Usage: go run main.go start server")
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func runServer() {
	fmt.Println("Server running on http://localhost:9000")
	http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
