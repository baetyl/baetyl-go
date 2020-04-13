package http

import (
	"github.com/baetyl/baetyl-go/log"
	"github.com/valyala/fasthttp"
)

// NewServer new server
func NewServer(cfg ServerConfig, handler fasthttp.RequestHandler) {
	server := fasthttp.Server{
		Handler:            handler,
		Concurrency:        cfg.Concurrency,
		DisableKeepalive:   cfg.DisableKeepalive,
		TCPKeepalive:       cfg.TCPKeepalive,
		MaxRequestBodySize: cfg.MaxRequestBodySize,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		IdleTimeout:        cfg.IdleTimeout,
	}

	go func() {
		logger := log.With(log.Any("http", "server"))
		logger.Info("server is running", log.Any("address", cfg.Address))
		if cfg.Cert != "" || cfg.Key != "" {
			if err := server.ListenAndServeTLS(cfg.Address, cfg.Cert, cfg.Key); err != nil {
				logger.Error("https server shutdown", log.Error(err))
			}
		} else {
			if err := server.ListenAndServe(cfg.Address); err != nil {
				logger.Error("http server shutdown", log.Error(err))
			}
		}
	}()
}
