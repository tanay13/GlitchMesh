package metrics

type Metrics interface {
	Increment(metric string, value int)
	Set(metric string, value int)
	Get() map[string]int
}
