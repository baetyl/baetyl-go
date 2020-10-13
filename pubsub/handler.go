package pubsub

type Handler interface {
	OnMessage(interface{}) error
	OnTimeout() error
}
