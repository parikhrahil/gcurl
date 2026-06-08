package transport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
)

type ExecutionResult struct {
	StatusCode       int
	BytesTransmitted int64
	BytesReceived    int64
	Duration         time.Duration
	Error            error
}

type ParallelEngine struct {
	cfg    *config.RequestConfiguration
	client *http.Client
}

func NewParallelEngine(cfg *config.RequestConfiguration) (*ParallelEngine, error) {
	client, err := NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ParallelEngine{
		cfg:    cfg,
		client: client,
	}, nil
}

func (p *ParallelEngine) Execute(ctx context.Context) ([]ExecutionResult, config.AuditMetrics) {
	var wg sync.WaitGroup

	jobs := make(chan int, p.cfg.TotalRequests)
	results := make(chan ExecutionResult, p.cfg.TotalRequests)

	for w := 0; w < p.cfg.Concurrency; w++ {
		wg.Add(1)
		go p.worker(ctx, &wg, jobs, results)
	}

	for j := 0; j < p.cfg.TotalRequests; j++ {
		jobs <- j
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var summary []ExecutionResult
	var globalMetrics config.AuditMetrics

	for res := range results {
		summary = append(summary, res)
		if res.Error == nil {
			globalMetrics.BytesTransmitted += res.BytesTransmitted
			globalMetrics.BytesReceived += res.BytesReceived
		}
	}
	return summary, globalMetrics
}

func (p *ParallelEngine) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan int,
	results chan<- ExecutionResult,
) {
	defer wg.Done()
	for range jobs {
		results <- p.execute(ctx)
	}
}

func (p *ParallelEngine) execute(ctx context.Context) ExecutionResult {
	var transmitted, received int64
	startTime := time.Now()

	var bodyReader io.Reader
	if p.cfg.Data != "" {
		bodyReader = strings.NewReader(p.cfg.Data)
	}

	req, err := http.NewRequestWithContext(ctx, p.cfg.Method, p.cfg.URL, bodyReader)
	if err != nil {
		return ExecutionResult{Error: err}
	}

	for k, v := range p.cfg.Headers {
		for _, val := range v {
			req.Header.Add(k, val)
		}
	}

	res, err := p.client.Do(req)
	if err != nil {
		return ExecutionResult{Error: err}
	}

	if res.Body != nil {
		defer res.Body.Close()
		n, _ := io.Copy(io.Discard, res.Body)
		received = n
	}

	if p.cfg.Data != "" {
		transmitted = int64(len(p.cfg.Data))
	}

	return ExecutionResult{
		StatusCode:       res.StatusCode,
		BytesTransmitted: transmitted,
		BytesReceived:    received,
		Duration:         time.Since(startTime),
		Error:            nil,
	}
}
