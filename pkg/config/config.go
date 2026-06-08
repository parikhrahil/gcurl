package config

import (
	"net/http"
	"time"
)

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
}

func NewDefaultConfig() *RequestConfiguration {
	return &RequestConfiguration{
		Method:         http.MethodGet,
		Headers:        make(http.Header),
		ConnectTimeout: 10 * time.Second,
		MaxTimeout:     30 * time.Second,
		Insecure:       false,
		Verbose:        false,
	}
}
