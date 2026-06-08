package transport

import (
	"net"
	"time"
)

func NewDialer(connectTimeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout:   connectTimeout,
		KeepAlive: 30 * time.Second,
	}
}
