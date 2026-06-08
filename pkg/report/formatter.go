package report

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
	"github.com/parikhrahil/gcurl/pkg/transport"
)

func RenderSummary(
	results []transport.ExecutionResult,
	wallclockTime time.Duration,
	metrics config.AuditMetrics,
) {
	totalRequests := len(results)
	if totalRequests == 0 {
		fmt.Fprintln(os.Stderr, "Error: No benchmark tracking data compiled.")
		return
	}

	var success, failures int
	statusDistribution := make(map[int]int)
	var validDurations []time.Duration

	for _, res := range results {
		if res.Error != nil || res.StatusCode >= 400 {
			failures++
		} else {
			success++
		}
		statusDistribution[res.StatusCode]++
		if res.Error == nil {
			validDurations = append(validDurations, res.Duration)
		}
	}

	requestsPerSecond := float64(totalRequests) / wallclockTime.Seconds()
	megaBytesReceived := float64(metrics.BytesReceived) / (1024 * 1024)
	mbPerSecond := megaBytesReceived / wallclockTime.Seconds()

	var avgTime, p50, p90, p99 time.Duration
	if len(validDurations) > 0 {
		sort.Slice(validDurations, func(i, j int) bool {
			return validDurations[i] < validDurations[j]
		})

		var totalDuration time.Duration
		for _, d := range validDurations {
			totalDuration += d
		}
		avgTime = totalDuration / time.Duration(len(validDurations))
		p50 = validDurations[int(float64(len(validDurations))*0.50)]
		p90 = validDurations[int(float64(len(validDurations))*0.90)]
		p99 = validDurations[int(float64(len(validDurations))*0.99)]
	}

	// Step 4: Stream the clean metric layout straight to the terminal standard output
	fmt.Println("\n=====================================================================")
	fmt.Println("                       gcurl BENCHMARK REPORT                        ")
	fmt.Println("=====================================================================")
	fmt.Printf("Total Elapsed Time (Wall-Clock):   %v\n", wallclockTime.Round(time.Millisecond))
	fmt.Printf("Completed Requests:                %d\n", totalRequests)
	fmt.Printf("Successful Transactions (2xx/3xx): %d\n", success)
	fmt.Printf("Failed Transactions (4xx/5xx/Net): %d\n", failures)
	fmt.Printf("System Throughput Efficiency:      %.2f requests/sec\n", requestsPerSecond)
	fmt.Printf("Network Transfer Volume:           %.3f MB total (%.2f MB/sec)\n", megaBytesReceived, mbPerSecond)
	fmt.Println("---------------------------------------------------------------------")
	fmt.Println("LATENCY PERCENTILE DISTRIBUTION:")
	fmt.Printf("  Average (Mean) Response Time:    %v\n", avgTime.Round(time.Microsecond))
	fmt.Printf("  50th Percentile (Median - p50):  %v\n", p50.Round(time.Microsecond))
	fmt.Printf("  90th Percentile (Tail - p90):    %v\n", p90.Round(time.Microsecond))
	fmt.Printf("  99th Percentile (Spike - p99):   %v\n", p99.Round(time.Microsecond))
	fmt.Println("---------------------------------------------------------------------")
	fmt.Println("HTTP STATUS CODE FREQUENCY DISTRIBUTION:")
	for code, freq := range statusDistribution {
		statusLabel := "OK"
		if code == 0 {
			statusLabel = "NETWORK_ERROR"
		} else if code >= 400 {
			statusLabel = "ERR"
		}
		fmt.Printf("  [HTTP %03d (%s)]: %d requests\n", code, statusLabel, freq)
	}
	fmt.Println("=====================================================================")
}
