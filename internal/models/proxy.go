package models

type Proxy struct {
	Service []ServiceConfig
}

type ServiceConfig struct {
	Name  string
	Url   string
	Fault string
	Value string
}
