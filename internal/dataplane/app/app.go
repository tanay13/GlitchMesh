package app

import (
	"log"
	"os"

	"github.com/tanay13/GlitchMesh/internal/dataplane/faults"
	"github.com/tanay13/GlitchMesh/internal/dataplane/proxy"
)

type App struct {
	ProxyService *proxy.ProxyService
	FaultService *faults.FaultService
	Logger       *log.Logger
}

func NewApp() *App {

	logger := log.New(os.Stdout, "[GlitchMesh] ", log.LstdFlags)

	faultInjector := &faults.FaultInjector{}

	faultService := faults.NewFaultService(faultInjector, logger)
	proxyService := proxy.NewProxyService(faultService, logger)

	return &App{
		proxyService,
		faultService,
		logger,
	}

}
