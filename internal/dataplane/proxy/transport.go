package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/tanay13/GlitchMesh/internal/dataplane/faults"
	"github.com/tanay13/GlitchMesh/internal/shared/constants"
	"github.com/tanay13/GlitchMesh/internal/shared/utils"
)

var (
	dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	SharedTransport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}

			// Intercept connection and register if service name context key is present
			if serviceName, ok := ctx.Value(constants.SERVICE_NAME_CTX_KEY).(string); ok {
				wrapped := faults.Registry.Register(serviceName, conn)
				return wrapped, nil
			}

			return conn, nil
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	sharedClient = &http.Client{
		Transport: SharedTransport,
	}
)

func ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) (int, error) {
	log.Printf("[proxy] forwarding %s %s", r.Method, targetURL)

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	resp, err := sharedClient.Do(req)

	if err != nil {
		return http.StatusBadGateway, err
	}

	defer resp.Body.Close()

	utils.CopyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("response body copy failed: %v", err)
	}

	return http.StatusOK, nil

}
