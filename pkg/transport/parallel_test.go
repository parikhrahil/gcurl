package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
	"github.com/parikhrahil/gcurl/pkg/transport"
	"github.com/stretchr/testify/require"
)

func TestParallelEngine_Execute_DeterministicDraining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	defer server.Close()

	cfg := &config.RequestConfiguration{
		Method:        http.MethodGet,
		URL:           server.URL,
		Headers:       make(http.Header),
		Concurrency:   3,
		TotalRequests: 10,
	}

	engine, err := transport.NewParallelEngine(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results, globalMetrics := engine.Execute(ctx)
	if len(results) != cfg.TotalRequests {
		t.Fatalf("Invariant Violation: Engine completed %d requests, expected exactly %d",
			len(results), cfg.TotalRequests)
	}

	var successfulRuns int
	for _, res := range results {
		if res.Error != nil {
			t.Errorf("Unexpected client network error recorded: %v", res.Error)
		}
		if res.StatusCode == http.StatusOK {
			successfulRuns++
		}
	}

	if successfulRuns != cfg.TotalRequests {
		t.Errorf("Expected %d successful status codes, but got %d", cfg.TotalRequests, successfulRuns)
	}

	// Validate that our byte counting decorator accurately captured payload telemetry
	expectedBytesReceived := int64(cfg.TotalRequests * 4) // "pong" is 4 bytes
	if globalMetrics.BytesReceived != expectedBytesReceived {
		t.Errorf("Telemetry Skew: Tracked %d bytes received, expected %d bytes",
			globalMetrics.BytesReceived, expectedBytesReceived)
	}
}
