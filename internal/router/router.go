package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tanay13/GlitchMesh/internal/app"
	"github.com/tanay13/GlitchMesh/internal/client"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

var appInstance *app.App

func InitRouter() {
	appInstance = app.NewApp()
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from GlitchMesh!"))
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/redirect/")
	urlParts := strings.SplitN(path, "/", 2)
	targetService := urlParts[0]

	start := time.Now()
	response, err := appInstance.ProxyService.HandleRequest(context.Background(), urlParts)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[Target: %s, Time Taken: %s , error: %v]", targetService, elapsed, err.Error())
		utils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if response.ShouldTerminate {

		errorMsg := response.Message
		if response.ContextErr != nil {
			errorMsg = fmt.Sprintf("%s (Context: %v)", response.Message, response.ContextErr)
		}

		utils.WriteJSONError(w, response.StatusCode, errorMsg)

		log.Printf("[Target: %s, Time Taken: %s, Fault: %s]", targetService, elapsed, response.Message)
		return
	}

	targetUrl := response.TargetUrl

	if r.URL.RawQuery != "" {
		targetUrl += "?" + r.URL.RawQuery
	}

	client.ProxyRequest(w, r, targetUrl)

	log.Printf("[Target: %s, Time Taken: %s]", targetService, elapsed)
}
