package models

type Metrics struct {
	FaultMetrics FaultsMetricsSnapshot `json:"fault_metrics"`
}

type SystemMetricsSnapshot struct {
	TotalRequest int `json:"total_requests"`
}

type FaultsMetricsSnapshot struct {
	TotalFaultsInjected int `json:"total_injected"`
	LatencyCount        int `json:"latency_count"`
	TimeoutCount        int `json:"timeout_count"`
	ErrorCount          int `json:"error_count"`
}
