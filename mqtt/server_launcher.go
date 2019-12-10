package mqtt

import (
	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl-go/utils"
)

// The Launcher helps with launching a server and accepting connections.
type Launcher struct {
	*transport.Launcher
}

// NewLauncher returns a new Launcher.
func NewLauncher(c utils.Certificate) (*Launcher, error) {
	lc := transport.LaunchConfig{}
	if c.Key != "" || c.Cert != "" {
		var err error
		lc.TLSConfig, err = utils.NewTLSConfigServer(c)
		if err != nil {
			return nil, err
		}
	}
	return &Launcher{Launcher: transport.NewLauncher(lc)}, nil
}
