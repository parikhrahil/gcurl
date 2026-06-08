package report_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
	"github.com/parikhrahil/gcurl/pkg/report"
	"github.com/parikhrahil/gcurl/pkg/transport"
)

func TestRenderSummary_PercentileDistributionAccuracy(t *testing.T) {
	results := []transport.ExecutionResult{
		{Duration: 10 * time.Millisecond, StatusCode: 200},
		{Duration: 11 * time.Millisecond, StatusCode: 200},
		{Duration: 12 * time.Millisecond, StatusCode: 200},
		{Duration: 12 * time.Millisecond, StatusCode: 200},
		{Duration: 13 * time.Millisecond, StatusCode: 200}, // Index 4
		{Duration: 14 * time.Millisecond, StatusCode: 200}, // Index 5 -> Expected p50
		{Duration: 15 * time.Millisecond, StatusCode: 200},
		{Duration: 16 * time.Millisecond, StatusCode: 200},
		{Duration: 100 * time.Millisecond, StatusCode: 404},
		{Duration: 500 * time.Millisecond, StatusCode: 500}, // Index 9 -> Expected p90/p99
	}

	wallClockTime := 200 * time.Millisecond
	globalMetrics := config.AuditMetrics{BytesReceived: 1024, BytesTransmitted: 256}

	// Intercept os.Stdout text streams to inspect string outputs
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the analytics renderer
	report.RenderSummary(results, wallClockTime, globalMetrics)

	// Restore the original standard output hooks
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	outputStr := buf.String()

	if !strings.Contains(outputStr, "14ms") {
		t.Errorf(
			"p50 Median Calculation Error: Expected 14ms distribution marker to display. Output: %s",
			outputStr,
		)
	}
	if !strings.Contains(outputStr, "500ms") {
		t.Errorf(
			"p90 Tail Calculation Error: Expected 500ms tail outlier marker to display. Output: %s",
			outputStr,
		)
	}
	if !strings.Contains(outputStr, "Completed Requests:                10") {
		t.Errorf("Volumetric Display Corrupted. Output: %s", outputStr)
	}
}
