package router

import (
	"log"
	"net/http"
	"strings"

	"github.com/tanay13/GlitchMesh/internal/logic"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func RedirectRequest(w http.ResponseWriter, r *http.Request) {
	endpoint := strings.TrimPrefix(r.URL.Path, "/redirect/")

	resp, err := logic.Redirect(endpoint)
	if err != nil {
		log.Println("Something went wrong", err)
		utils.WriteJSONError(w, http.StatusBadGateway, "Failed to redirect the request")
		return
	}

	w.Write(resp)
}
