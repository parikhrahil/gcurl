package config

import (
	"net/http"
	"time"
)

type AuditMetrics struct {
	BytesTransmitted       int64
	BytesReceived          int64
	DNSLookupDuration      time.Duration
	TCPHandshakeDuration   time.Duration
	TLSTerminationDuration time.Duration
	TotalDuration          time.Duration
}

// RequestConfiguration defines the immutable domain model for a network request.
type RequestConfiguration struct {
	Method         string
	URL            string
	Headers        http.Header
	Data           string
	ConnectTimeout time.Duration
	MaxTimeout     time.Duration
	Insecure       bool
	Verbose        bool
	Metrics        AuditMetrics
	Concurrency    int
	TotalRequests  int
	CACertPath     string
	TLSMinVer      string
}

func NewDefaultConfig() *RequestConfiguration {
	return &RequestConfiguration{
		Method:         http.MethodGet,
		Headers:        make(http.Header),
		ConnectTimeout: 10 * time.Second,
		MaxTimeout:     30 * time.Second,
		Insecure:       false,
		Verbose:        false,
		Concurrency:    1,
		TotalRequests:  1,
	}
}
