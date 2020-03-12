package http

import (
	"crypto/tls"
	"time"
)

// ClientOptions link client options
type ClientOptions struct {
	Address               string
	TLSConfig             *tls.Config
	Timeout               time.Duration
	KeepAlive             time.Duration
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// NewClientOptions creates link client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		Timeout:               30 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
