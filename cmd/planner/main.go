package main

import (
	"fmt"

	"github.com/tanay13/GlitchMesh/internal/controlplane/planner"
)

func main() {
	err := planner.PlannerInit()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	planner.Planner()
}
