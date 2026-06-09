package transport

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !cfg.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) > cfg.MaxRedirects {
				return fmt.Errorf("stopped after 10 consecutive redirects to prevent infinite loop vectors")
			}
			return nil
		},
	}, nil
}

func NewHTTPTrace(w io.Writer, cfg *config.RequestConfiguration) *httptrace.ClientTrace {
	var protocol string
	var host string
	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			host = info.Host
			fmt.Fprintf(w, "* Resolving host: %s\n", info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if info.Err != nil {
				fmt.Fprintf(w, "* Could not resolve host: %s\n", host)
				fmt.Fprintf(w, "* Closing connection\n")
				return
			}
			ipAddr := make([]string, len(info.Addrs))
			for i := 0; i < len(info.Addrs); i++ {
				ipAddr[i] = info.Addrs[i].String()
			}
			fmt.Fprintf(w, "* Resolved IP: %s\n", strings.Join(ipAddr, ", "))
		},
		ConnectStart: func(network, addr string) {
			fmt.Fprintf(w, "* Trying %s...\n", addr)
		},
		ConnectDone: func(network, addr string, err error) {
			if err == nil {
				fmt.Fprintf(w, "* Connected to %s\n", addr)
			}
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err == nil {
				dateFormat := "Jan 2 15:04:05 2006 GMT"
				startDate := state.PeerCertificates[0].NotBefore.Format(dateFormat)
				expiryDate := state.PeerCertificates[0].NotAfter.Format(dateFormat)
				if strings.Contains(state.NegotiatedProtocol, "h2") {
					protocol = "HTTP/2"
				} else {
					protocol = "HTTP/1.1"
				}
				fmt.Fprintf(w, "* SSL Handshake successful using %s\n",
					tls.VersionName(state.Version))
				fmt.Fprintf(w, "* ALPN: server accepted %s\n", state.NegotiatedProtocol)
				fmt.Fprintf(w, "* Server Certificate:\n")
				fmt.Fprintf(w, "*  subject: %s\n", state.PeerCertificates[0].Subject)
				fmt.Fprintf(w, "*  start date: %s\n", startDate)
				fmt.Fprintf(w, "*  expiry date: %s\n", expiryDate)
				fmt.Fprintf(w, "*  subjectAltName: %s\n", strings.Join(state.PeerCertificates[0].DNSNames, "; "))
				fmt.Fprintf(w, "*  issuer: %s\n", state.PeerCertificates[0].Issuer.String())
				fmt.Fprintf(w, "*  SSL certificate verify ok.\n")
				fmt.Fprintf(w, "* using %s\n", protocol)
			}
		},
		WroteRequest: func(wri httptrace.WroteRequestInfo) {
			fmt.Fprintf(w, "> %s %s %s\n", cfg.Method, cfg.URL, protocol)
			fmt.Fprintf(w, ">\n")
		},
		WroteHeaderField: func(key string, value []string) {
			valStr := strings.Join(value, ", ")
			if strings.HasPrefix(key, ":") {
				fmt.Fprintf(w, "* [%s] [1] [%s: %s]\n", protocol, key, valStr)
			} else {
				fmt.Fprintf(w, "> %s: %s\n", key, valStr)
			}
		},
	}
}
