package logic

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/tanay13/GlitchMesh/internal/client"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func ProxyLogic(w http.ResponseWriter, r *http.Request, urlParts []string) (int, error) {

	proxyConfig, err := utils.ParseConfigYaml()

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("%v", err)
	}

	serviceName, endpoint := utils.ParseURLParts(urlParts)

	serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)

	if serviceConfig == nil {
		return http.StatusInternalServerError, fmt.Errorf("service config not found in the proxy list")
	}

	serviceURL, err := url.Parse(serviceConfig.Url)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	serviceURL.Path = endpoint
	serviceURL.RawQuery = r.URL.RawQuery

	FaultInjection(serviceConfig.Fault)

	statusCode, err := client.ProxyRequest(w, r, serviceURL.String())

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return statusCode, nil
}
