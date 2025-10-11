package service

import (
	"context"
	"fmt"
	"log"

	"github.com/tanay13/GlitchMesh/internal/config"
	domain "github.com/tanay13/GlitchMesh/internal/domain/faults"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

type ProxyService struct {
	faultService *FaultService
	logger       *log.Logger
}

func NewProxyService(faultService *FaultService, logger *log.Logger) *ProxyService {
	return &ProxyService{
		faultService,
		logger,
	}
}

func (s *ProxyService) HandleRequest(ctx context.Context, urlParts []string) (*domain.FaultResponse, error) {

	/* remove parsing everytime there is a request, better way is to store it and use it again and again or hot-reloading */
	proxyConfig := config.ProxyConfig

	serviceName, endpoint := utils.ParseURLParts(urlParts)

	serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)

	targetUrl := serviceConfig.Url + endpoint

	if serviceConfig == nil {
		return nil, fmt.Errorf("service config not found in the proxy list")
	}

	faultResponse := s.faultService.ApplyFault(ctx, serviceConfig.Fault)

	faultResponse.TargetUrl = targetUrl

	return faultResponse, nil
}
