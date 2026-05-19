package main

import (
	"fmt"

	"github.com/tanay13/GlitchMesh/internal/controlplan"
)

func main() {
	err := controlplan.PlannerInit()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	controlplan.Planner()
}
