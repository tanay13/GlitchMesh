package trafficgen

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

func GenerateTraffic(options models.Options) {

	concurrency := options.Concurrency
	count := options.Count
	timeout := options.Timeout
	url := options.Url

	if url == "" {
		log.Fatal("Url not passed in traffic generation command")
	}

	if concurrency < 1 {
		log.Fatal("concurrency must be >= 1")
	}
	if count < 1 {
		log.Fatal("count must be >= 1")
	}

	client := &http.Client{Timeout: timeout}

	var okCount, errCount atomic.Int64
	jobs := make(chan int, count)
	var wg sync.WaitGroup

	log.Printf("[traffic-gen] url=%s concurrency=%d count=%d", url, concurrency, count)
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for id := range jobs {
				reqStart := time.Now()
				resp, err := client.Get(url)
				elapsed := time.Since(reqStart)

				if err != nil {
					errCount.Add(1)
					log.Printf("[traffic-gen] worker=%d req=%d error=%v elapsed=%s", workerID, id, err, elapsed)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					okCount.Add(1)
					log.Printf("[traffic-gen] worker=%d req=%d status=%d elapsed=%s", workerID, id, resp.StatusCode, elapsed)
				} else {
					errCount.Add(1)
					log.Printf("[traffic-gen] worker=%d req=%d status=%d elapsed=%s", workerID, id, resp.StatusCode, elapsed)
				}
			}
		}(i + 1)
	}

	for i := 1; i <= count; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	total := time.Since(start)
	fmt.Printf("\n--- traffic-gen summary ---\n")
	fmt.Printf("url:         %s\n", url)
	fmt.Printf("total:       %d\n", count)
	fmt.Printf("success:     %d\n", okCount.Load())
	fmt.Printf("errors:      %d\n", errCount.Load())
	fmt.Printf("duration:    %s\n", total)
	fmt.Printf("req/sec:     %.2f\n", float64(count)/total.Seconds())

	if errCount.Load() > 0 {
		os.Exit(1)
	}
}
