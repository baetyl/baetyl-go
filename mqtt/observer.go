package mqtt

import "github.com/256dpi/gomqtt/packet"

// OnPublish handles publish packet
type OnPublish func(*packet.Publish) error

// OnPuback handles puback packet
type OnPuback func(*packet.Puback) error

// OnError handles error
type OnError func(error)

// Observer message observer interface
type Observer interface {
	OnPublish(*packet.Publish) error
	OnPuback(*packet.Puback) error
	// can't invoke client.Close() or will cause deadlock
	OnError(error)
}

// ObserverWrapper MQTT message handler wrapper
type ObserverWrapper struct {
	onPublish OnPublish
	onPuback  OnPuback
	onError   OnError
}

// NewObserverWrapper creates a new handler wrapper
func NewObserverWrapper(onPublish OnPublish, onPuback OnPuback, onError OnError) *ObserverWrapper {
	return &ObserverWrapper{
		onPublish: onPublish,
		onPuback:  onPuback,
		onError:   onError,
	}
}

// OnPublish handles publish packet
func (h *ObserverWrapper) OnPublish(pkt *packet.Publish) error {
	if h.onPublish == nil {
		return nil
	}
	return h.onPublish(pkt)
}

// OnPuback handles puback packet
func (h *ObserverWrapper) OnPuback(pkt *packet.Puback) error {
	if h.onPuback == nil {
		return nil
	}
	return h.onPuback(pkt)
}

// OnError handles error
func (h *ObserverWrapper) OnError(err error) {
	if h.onError == nil {
		return
	}
	h.onError(err)
}
