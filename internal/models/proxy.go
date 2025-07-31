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
	Type  string
	Value int
}
