package audit

import (
	"fmt"
	"io"
	"time"
)

type AuditReader struct {
	underlying io.Reader
	byteCount  *int64
}

type ProgressReader struct {
	underlying  io.Reader
	bytesCount  *int64
	totalBytes  int64
	startTime   time.Time
	lastUpdated time.Time
	errSink     io.Writer
	active      bool
}

func NewAuditReader(r io.Reader, count *int64) io.Reader {
	return &AuditReader{
		underlying: r,
		byteCount:  count,
	}
}

func (r *AuditReader) Read(p []byte) (int, error) {
	n, err := r.underlying.Read(p)
	if n > 0 {
		*r.byteCount += int64(n)
	}
	return n, err
}

func NewProgressReader(r io.Reader, counter *int64, bytes int64, errSink io.Writer, active bool) *ProgressReader {
	now := time.Now()
	return &ProgressReader{
		underlying:  r,
		bytesCount:  counter,
		totalBytes:  bytes,
		startTime:   now,
		lastUpdated: now,
		errSink:     errSink,
		active:      active,
	}
}

func (r *ProgressReader) Read(p []byte) (int, error) {
	n, err := r.underlying.Read(p)
	if n > 0 {
		*r.bytesCount += int64(n)
		r.renderProgress(false)
	}
	if err == io.EOF && r.active {
		r.renderProgress(true)
	}
	return n, err
}

func (r *ProgressReader) renderProgress(isFinal bool) {
	if !r.active {
		return
	}
	now := time.Now()
	// Throttling Enforcement: Only redraw if 100ms elapsed, or if this is the final block
	if !isFinal && now.Sub(r.lastUpdated) < 100*time.Millisecond {
		return
	}

	r.lastUpdated = now

	currentRead := *r.bytesCount
	elapsed := now.Sub(r.startTime).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}

	speedBitsPerSec := float64(currentRead) / elapsed
	speedHumanReadable := formatBytes(int64(speedBitsPerSec))

	if r.totalBytes > 0 {
		percent := (float64(currentRead) / float64(r.totalBytes)) * 100
		bytesRemaining := r.totalBytes - currentRead
		etaSeconds := int(float64(bytesRemaining) / speedBitsPerSec)
		etaStr := (time.Duration(etaSeconds) * time.Second).String()
		fmt.Fprintf(r.errSink, "\r* Progress: %3.1f%% | %s / %s | %s/s | ETA: %s   ",
			percent, formatBytes(currentRead), formatBytes(r.totalBytes), speedHumanReadable, etaStr)
	} else {
		fmt.Fprintf(r.errSink, "\r* Progress: Streaming... | %s transferred | %s/s   ",
			formatBytes(currentRead), speedHumanReadable)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
