package link

// OnMsg handles next message
type OnMsg func(*Message) error

// OnAck handles message ack
type OnAck func(*Message) error

// OnErr handles error
type OnErr func(error)

// Observer message observer interface
type Observer interface {
	OnMsg(*Message) error
	OnAck(*Message) error
	OnErr(error)
}

// // ObserverWrapper MQTT message handler wrapper
// type ObserverWrapper struct {
// 	onMsg OnMsg
// 	onAck OnAck
// 	onErr OnErr
// }

// // NewObserverWrapper creates a new handler wrapper
// func NewObserverWrapper(onMsg OnMsg, onAck OnAck, onErr OnErr) *ObserverWrapper {
// 	return &ObserverWrapper{
// 		onMsg: onMsg,
// 		onAck: onAck,
// 		onErr: onErr,
// 	}
// }

// // OnMsg handles next message
// func (h *ObserverWrapper) OnMsg(pkt *Message) error {
// 	if h.onMsg == nil {
// 		return nil
// 	}
// 	return h.onMsg(pkt)
// }

// // OnAck handles message ack
// func (h *ObserverWrapper) OnAck(pkt *Message) error {
// 	if h.onAck == nil {
// 		return nil
// 	}
// 	return h.onAck(pkt)
// }

// // OnErr handles error
// func (h *ObserverWrapper) OnErr(err error) {
// 	if h.onErr == nil {
// 		return
// 	}
// 	h.onErr(err)
// }
