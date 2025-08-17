package router

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tanay13/GlitchMesh/internal/logic"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/redirect/")
	urlParts := strings.SplitN(path, "/", 2)
	targetService := urlParts[0]

	start := time.Now()
	statusCode, faultType, faultValue, err := service.HandleRequest(urlParts)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[Target: %s, Time Taken: %s , error: %v]", targetService, elapsed, err.Error())
		utils.WriteJSONError(w, statusCode, err.Error())
		return
	}

	log.Printf("[Target: %s, Time Taken: %s, Fault: %s, Value: %s]", targetService, elapsed, faultType, faultValue)
}
