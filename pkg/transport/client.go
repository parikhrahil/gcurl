package transport

import (
	"net/http"
	"time"

	"github.com/parikhrahil/gcurl/pkg/config"
)

func NewHTTPClient(cfg *config.RequestConfiguration) (*http.Client, error) {
	dialer := NewDialer(cfg.ConnectTimeout)
	tlsConfig, err := BuildTlsConfig(cfg.CACertPath, cfg.TLSMinVer)
	if err != nil {
		return nil, err
	}
	tlsConfig.InsecureSkipVerify = cfg.Insecure

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSClientConfig:       tlsConfig,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   cfg.Concurrency + 2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.MaxTimeout,
	}, nil
}
