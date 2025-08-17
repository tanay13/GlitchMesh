package service

import (
	"fmt"
	"net/http"

	domain "github.com/tanay13/GlitchMesh/internal/domain/faults"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

type ProxyService struct {
	faultService *FaultService
}

func (s *ProxyService) HandleRequest(urlParts []string) *domain.FaultResponse {

	/* remove parsing everytime there is a request, better way is to store it and use it again and again or hot-reloading */
	proxyConfig, err := utils.ParseConfigYaml()

	if err != nil {
		return &domain.FaultResponse{
			Applied:         false,
			ShouldTerminate: true,
			StatusCode:      http.StatusInternalServerError,
			Body:            fmt.Errorf("%v", err),
		}
	}

	serviceName, _ := utils.ParseURLParts(urlParts)

	serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)

	if serviceConfig == nil {
		return &domain.FaultResponse{
			Applied:         false,
			ShouldTerminate: true,
			StatusCode:      http.StatusInternalServerError,
			Body:            fmt.Errorf("service config not found in the proxy list"),
		}
	}

	faultResponse := s.faultService.ApplyFault(serviceConfig.Fault)

	return faultResponse
}
