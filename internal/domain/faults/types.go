package domain

type FaultResponse struct {
	Applied         bool
	ShouldTerminate bool
	StatusCode      int
	Message         string
	Body            any
}

type Fault interface {
	InjectFault() FaultResponse
}
