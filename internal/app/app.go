package app

import (
	"log"
	"os"

	domain "github.com/tanay13/GlitchMesh/internal/domain/faults"
	"github.com/tanay13/GlitchMesh/internal/domain/service"
)

type App struct {
	ProxyService *service.ProxyService
	FaultService *service.FaultService
	Logger       *log.Logger
}

func NewApp() *App {

	logger := log.New(os.Stdout, "[GlitchMesh] ", log.LstdFlags)

	faultInjector := &domain.FaultInjector{
		IsFaultEnabled: false,
		FaultsEnabled:  make(map[domain.Fault]any),
	}

	faultService := service.NewFaultService(faultInjector, logger)
	proxyService := service.NewProxyService(faultService, logger)

	return &App{
		proxyService,
		faultService,
		logger,
	}

}
