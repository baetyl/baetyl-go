package mqtt

import (
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// Handle handles connection
type Handle func(Connection, bool)

// Endpoint the endpoint
type Endpoint struct {
	Address   string
	Anonymous bool
	Handle    Handle
}

// Transport transport
type Transport struct {
	endpoints []*Endpoint
	servers   []Server
	log       *log.Logger
	utils.Tomb
}

// NewTransport creates a new transport
func NewTransport(endpoints []*Endpoint, cert utils.Certificate) (*Transport, error) {
	launcher, err := NewLauncher(cert)
	if err != nil {
		return nil, err
	}
	tp := &Transport{
		endpoints: endpoints,
		servers:   make([]Server, 0),
		log:       log.With(log.Any("transport", "mqtt")),
	}
	for _, endpoint := range endpoints {
		if endpoint.Handle == nil {
			panic("endpoint handle cannot be nil")
		}
		svr, err := launcher.Launch(endpoint.Address)
		if err != nil {
			tp.Close()
			return nil, err
		}
		tp.servers = append(tp.servers, svr)
		tp.accepting(svr, endpoint.Handle, endpoint.Anonymous)
	}
	tp.log.Info("transport has initialized")
	return tp, nil
}

func (tp *Transport) accepting(svr Server, handle Handle, anonymous bool) {
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
			handle(conn, anonymous)
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
