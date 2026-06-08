package transport

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
)

func NewHTTPClient(cfg *config.RequestConfiguration) *http.Client {
	dialer := NewDialer(cfg.ConnectTimeout)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSClientConfig:       tlsConfig,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.MaxTimeout,
	}
}
