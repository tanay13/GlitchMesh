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

	proxyConfig := config.GetProxyConfig()

	serviceName, endpoint, err := utils.ParseURLParts(urlParts)
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}

	serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)

	if serviceConfig == nil {
		return nil, fmt.Errorf("service '%s' not found in the proxy configuration", serviceName)
	}

	targetUrl := serviceConfig.Url + endpoint

	faultResponse := s.faultService.ApplyFault(ctx, serviceConfig.Fault)

	faultResponse.TargetUrl = targetUrl

	return faultResponse, nil
}
