package models

type Proxy struct {
	Service []ServiceConfig
}

type ServiceConfig struct {
	Name  string
	Url   string
	Fault Fault
}

type Fault struct {
	Enabled  bool
	Priority []string
	Types    map[string]FaultConfig
}

type FaultConfig struct {
	StatusCode int
	Message    string
	Delay      int
}
