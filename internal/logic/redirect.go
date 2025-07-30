package logic

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tanay13/GlitchMesh/internal/client"
	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func ProxyLogic(w http.ResponseWriter, r *http.Request) (int, error) {

	yamlFilePath := config.Configs.Env.YAML_FILE_PATH

	if yamlFilePath == "" {
		log.Fatalf("Please set up the proxy yaml first!!")
	}

	proxyConfig, err := utils.ParseYaml(yamlFilePath)

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("%v", err)
	}

	path := strings.TrimPrefix(r.URL.Path, "/redirect/")

	urlParts := strings.SplitN(path, "/", 2)

	serviceName := urlParts[0]

	endpoint := urlParts[1]

	serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)

	if serviceConfig == nil {
		return http.StatusInternalServerError, fmt.Errorf("Service config not found in the proxy list")
	}

	serviceUrl, err := url.Parse(serviceConfig.Url)

	if err != nil {
		return http.StatusInternalServerError, err
	}
	serviceUrl.Path = endpoint
	serviceUrl.RawQuery = r.URL.RawQuery

	statusCode, err := client.ProxyRequest(w, r, serviceUrl.String())

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return statusCode, nil
}
