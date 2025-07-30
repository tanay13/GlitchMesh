package client

import (
	"io"
	"log"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/utils"
)

func ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) (int, error) {

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return http.StatusBadGateway, err
	}

	defer resp.Body.Close()

	utils.CopyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("response body copy failed: %v", err)
	}

	utils.CopyHeaders(req.Header, r.Header)

	return http.StatusOK, nil

}
