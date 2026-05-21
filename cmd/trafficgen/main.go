package main

import (
	"flag"
	"time"

	"github.com/tanay13/GlitchMesh/internal/dataplane/trafficgen"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

func main() {

	url := flag.String("url", "", "target URL")
	concurrency := flag.Int("concurrency", 5, "number of concurrent workers")
	count := flag.Int("count", 50, "total requests to send")
	timeout := flag.Duration("timeout", 30*time.Second, "per-request timeout")
	flag.Parse()

	options := models.Options{
		Url:         *url,
		Concurrency: *concurrency,
		Count:       *count,
		Timeout:     *timeout,
	}

	trafficgen.GenerateTraffic(options)

}
