package proxy

import (
	"context"
	"fmt"
	"log"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/dataplane/faults"
	"github.com/tanay13/GlitchMesh/internal/shared/utils"
)

type ProxyService struct {
	faultService *faults.FaultService
	logger       *log.Logger
}

func NewProxyService(faultService *faults.FaultService, logger *log.Logger) *ProxyService {
	return &ProxyService{
		faultService,
		logger,
	}
}

func (s *ProxyService) HandleRequest(ctx context.Context, urlParts []string) (*faults.FaultResponse, error) {
	serviceName, endpoint, err := utils.ParseURLParts(urlParts)
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}

	serviceConfig := config.GetEffectiveServiceConfig(serviceName)

	if serviceConfig == nil {
		return nil, fmt.Errorf("service '%s' not found in the proxy configuration", serviceName)
	}

	targetUrl := serviceConfig.Url + endpoint

	faultResponse := s.faultService.ApplyFault(ctx, serviceConfig.Fault)

	faultResponse.TargetUrl = targetUrl

	return faultResponse, nil
}
