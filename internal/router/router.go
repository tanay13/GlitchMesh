package router

import (
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/logic"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {

	statusCode, err := logic.ProxyLogic(w, r)
	if err != nil {
		utils.WriteJSONError(w, statusCode, err.Error())
		return
	}

}
