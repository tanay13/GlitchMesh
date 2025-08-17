package models

type Proxy struct {
	Service []ServiceConfig
}

type ServiceConfig struct {
	Name  string
	Url   string
	Fault FaultConfig
}

type FaultConfig struct {
	Enabled bool
	Error   *ErrorFaultConfig
	Latency *LatencyFaultConfig
}

type ErrorFaultConfig struct {
	StatusCode int
	Message    string
}

type LatencyFaultConfig struct {
	Delay int
}
