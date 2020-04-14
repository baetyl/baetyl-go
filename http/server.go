package http

import (
	"github.com/baetyl/baetyl-go/log"
	"github.com/valyala/fasthttp"
)

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
			if err := s.ListenAndServeTLS(s.conf.Address, s.conf.Cert, s.conf.Key); err != nil {
				logger.Error("https server shutdown", log.Error(err))
			}
		} else {
			if err := s.ListenAndServe(s.conf.Address); err != nil {
				logger.Error("http server shutdown", log.Error(err))
			}
		}
	}()
}

func (s *Server) Close() {
	if s.Server != nil {
		s.Server.Shutdown()
	}
}
