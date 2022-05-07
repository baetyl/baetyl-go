package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl-go/v2/log"
)

var errNoCertOrKeyProvided = errors.New("cert or key has not provided")

type Server struct {
	conf ServerConfig
	*fasthttp.Server
}

// NewServer new server
func NewServer(cfg ServerConfig, handler fasthttp.RequestHandler) *Server {
	return &Server{
		conf: cfg,
		Server: &fasthttp.Server{
			Handler:            handler,
			Concurrency:        cfg.Concurrency,
			DisableKeepalive:   cfg.DisableKeepalive,
			TCPKeepalive:       cfg.TCPKeepalive,
			MaxRequestBodySize: cfg.MaxRequestBodySize,
			ReadTimeout:        cfg.ReadTimeout,
			WriteTimeout:       cfg.WriteTimeout,
			IdleTimeout:        cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() {
	go func() {
		logger := log.With(log.Any("http", "server"))
		logger.Info("server is running", log.Any("address", s.conf.Address))
		if s.conf.Cert != "" || s.conf.Key != "" {
			if len(s.conf.CA) != 0 {
				if err := s.ListenAndServeMTLS(s.conf.Address, s.conf.Cert, s.conf.Key); err != nil {
					logger.Error("https server shutdown", log.Error(err))
				}
			} else {
				if err := s.ListenAndServeTLS(s.conf.Address, s.conf.Cert, s.conf.Key); err != nil {
					logger.Error("https server shutdown", log.Error(err))
				}
			}
		} else {
			if err := s.ListenAndServe(s.conf.Address); err != nil {
				logger.Error("http server shutdown", log.Error(err))
			}
		}
	}()
}

func (s *Server) ListenAndServeMTLS(addr, certFile, keyFile string) error {
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("cannot load TLS key pair from certFile=%q and keyFile=%q: %s", certFile, keyFile, err)
	}
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
	}
	if len(s.conf.CA) != 0 {
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(s.conf.CA)
		if err != nil {
			return fmt.Errorf("cannot load TLS ca from caFile=%q: %s", s.conf.CA, err)
		}
		pool.AppendCertsFromPEM(caCrt)
		tlsConfig.ClientAuth = s.conf.ClientAuthType
		tlsConfig.ClientCAs = pool
	}
	tlsConfig.BuildNameToCertificate()
	if s.TCPKeepalive {
		if tcpln, ok := ln.(*net.TCPListener); ok {
			return s.Serve(tls.NewListener(tcpKeepaliveListener{
				TCPListener:     tcpln,
				keepalivePeriod: s.TCPKeepalivePeriod,
			}, tlsConfig))
		}
	}
	return s.Serve(tls.NewListener(ln, tlsConfig))
}

func (s *Server) Close() {
	if s.Server != nil {
		s.Server.Shutdown()
	}
}

type tcpKeepaliveListener struct {
	*net.TCPListener
	keepalivePeriod time.Duration
}

func (ln tcpKeepaliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if err := tc.SetKeepAlive(true); err != nil {
		tc.Close() //nolint:errcheck
		return nil, err
	}
	if ln.keepalivePeriod > 0 {
		if err := tc.SetKeepAlivePeriod(ln.keepalivePeriod); err != nil {
			tc.Close() //nolint:errcheck
			return nil, err
		}
	}
	return tc, nil
}
