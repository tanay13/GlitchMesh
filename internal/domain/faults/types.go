package domain

import "context"

type FaultResponse struct {
	Applied         bool
	TargetUrl       string
	ShouldTerminate bool
	StatusCode      int
	Message         string
	Body            any
	ContextErr      error
}

type Fault interface {
	InjectFault(context.Context) FaultResponse
}
