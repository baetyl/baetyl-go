package mqtt

import (
	"crypto/tls"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// Handle handles connection
type Handle func(Connection)

// Transport transport
type Transport struct {
	servers []Server
	log     *log.Logger
	utils.Tomb
}

// NewTransport creates a new transport
func NewTransport(cfg ServerConfig, handle Handle) (*Transport, error) {
	if handle == nil {
		panic("mqtt transport handle cannot be nil")
	}
	var err error
	var tlsconf *tls.Config
	if cfg.Certificate.Key != "" || cfg.Certificate.Cert != "" {
		tlsconf, err = utils.NewTLSConfigServer(cfg.Certificate)
		if err != nil {
			return nil, err
		}
	}
	tp := &Transport{
		servers: make([]Server, 0),
		log:     log.With(log.Any("mqtt", "server")),
	}
	launcher := NewLauncher(tlsconf)
	for _, address := range cfg.Addresses {
		svr, err := launcher.Launch(address)
		if err != nil {
			tp.Close()
			return nil, err
		}
		tp.servers = append(tp.servers, svr)
		tp.accepting(svr, handle)
	}
	tp.log.Info("transport has initialized")
	return tp, nil
}

func (tp *Transport) accepting(svr Server, handle Handle) {
	tp.Go(func() error {
		l := log.With(log.Any("server", svr.Addr().String()))
		l.Info("server starts to accept")
		defer l.Info("server has stopped accepting")

		for {
			conn, err := svr.Accept()
			if err != nil {
				if !tp.Alive() {
					l.Debug("failed to accept connection", log.Error(err))
					return nil
				}
				l.Error("failed to accept connection", log.Error(err))
				return err
			}
			handle(conn)
		}
	})
}

// Close closes service
func (tp *Transport) Close() error {
	tp.log.Info("transport is closing")
	defer tp.log.Info("transport has closed")

	tp.Kill(nil)
	for _, svr := range tp.servers {
		svr.Close()
	}
	return tp.Wait()
}
