package mqtt

import (
	"crypto/tls"
	"errors"
	"net"
	"time"

	gomqtt "github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/client/future"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/topic"
	"github.com/256dpi/gomqtt/transport"
)

// The supported MQTT versions.
const (
	Version311 byte = 4
	Version31  byte = 3
)

// The ConnackCode represents the return code in a Connack packet.
type ConnackCode = packet.ConnackCode

// All available ConnackCodes.
const (
	ConnectionAccepted ConnackCode = iota
	InvalidProtocolVersion
	IdentifierRejected
	ServerUnavailable
	BadUsernameOrPassword
	NotAuthorized
)

// QOS the quality of service levels
type QOS = packet.QOS

// Al QoS levels
const (
	QOSAtMostOnce  QOS = iota
	QOSAtLeastOnce QOS = iota
	QOSExactlyOnce QOS = iota
	QOSFailure     QOS = 0x80
)

// Type the packet type
type Type = packet.Type

// ID the packet id
type ID = packet.ID

// Packet the generic packet
type Packet = packet.Generic

// Publish the publish packet
type Publish = packet.Publish

// NewPublish creates a new Publish packet
func NewPublish() *Publish {
	return &Publish{}
}

// Puback the puback packet
type Puback = packet.Puback

// NewPuback creates a new Puback packet
func NewPuback() *Puback {
	return &Puback{}
}

// Subscribe the subscribe packet
type Subscribe = packet.Subscribe

// NewSubscribe creates a new Subscribe packet
func NewSubscribe() *Subscribe {
	return &Subscribe{}
}

// Suback the suback packet
type Suback = packet.Suback

// NewSuback creates a new Suback packet
func NewSuback() *Suback {
	return &Suback{}
}

// Unsuback the unsuback packet
type Unsuback = packet.Unsuback

// NewUnsuback creates a new Unsuback packet
func NewUnsuback() *Unsuback {
	return &Unsuback{}
}

// Pingreq the pingreq packet
type Pingreq = packet.Pingreq

// NewPingreq creates a new Pingreq packet.
func NewPingreq() *Pingreq {
	return &Pingreq{}
}

// Pingresp the pingresp packet
type Pingresp = packet.Pingresp

// NewPingresp creates a new Pingresp packet
func NewPingresp() *Pingresp {
	return &Pingresp{}
}

// Unsubscribe the unsubscribe packet
type Unsubscribe = packet.Unsubscribe

// NewUnsubscribe creates a new Unsubscribe packet
func NewUnsubscribe() *Unsubscribe {
	return &Unsubscribe{}
}

// Connect the connect packet
type Connect = packet.Connect

// NewConnect creates a new Connect packet
func NewConnect() *Connect {
	return &Connect{
		CleanSession: true,
		Version:      4,
	}
}

// Connack the connack packet
type Connack = packet.Connack

// NewConnack creates a new Connack packet
func NewConnack() *Connack {
	return &Connack{}
}

// Disconnect the disconnect packet
type Disconnect = packet.Disconnect

// NewDisconnect creates a new Disconnect packet
func NewDisconnect() *Disconnect {
	return &Disconnect{}
}

// Subscription the topic and qos of subscription
type Subscription = packet.Subscription

// Counter the packet id counter
type Counter = session.IDCounter

// NewCounter creates a new counter
func NewCounter() *Counter {
	return session.NewIDCounter()
}

// Trie the trie of topic subscription
type Trie = topic.Tree

// NewTrie creates a new trie
func NewTrie() *Trie {
	return topic.NewStandardTree()
}

// Tracker keeps track of keep alive intervals
type Tracker = gomqtt.Tracker

// NewTracker creates a new tracker
func NewTracker(timeout time.Duration) *Tracker {
	return gomqtt.NewTracker(timeout)
}

// Future future
type Future = future.Future

// NewFuture creates a new future.
func NewFuture() *Future {
	return future.New()
}

// Server the server to accept connections
type Server = transport.Server

// Connection the connection between a client and a server
type Connection = transport.Conn

// IsBidirectionalAuthentication check bidirectional authentication
func IsBidirectionalAuthentication(conn Connection) bool {
	var inner net.Conn
	if nc, ok := conn.(*transport.NetConn); ok {
		inner = nc.UnderlyingConn()
	} else if wss, ok := conn.(*transport.WebSocketConn); ok {
		inner = wss.UnderlyingConn().UnderlyingConn()
	}
	tlsconn, ok := inner.(*tls.Conn)
	if !ok {
		return false
	}
	state := tlsconn.ConnectionState()
	if !state.HandshakeComplete {
		return false
	}
	return len(state.PeerCertificates) > 0
}

// all gomqtt client errors
var (
	// client's erros
	ErrClientAlreadyConnecting  = gomqtt.ErrClientAlreadyConnecting
	ErrClientNotConnected       = gomqtt.ErrClientNotConnected
	ErrClientMissingID          = gomqtt.ErrClientMissingID
	ErrClientConnectionDenied   = gomqtt.ErrClientConnectionDenied
	ErrClientMissingPong        = gomqtt.ErrClientMissingPong
	ErrClientExpectedConnack    = gomqtt.ErrClientExpectedConnack
	ErrClientSubscriptionFailed = gomqtt.ErrFailedSubscription
	ErrClientAlreadyClosed      = errors.New("client is closed")

	// future's errors
	ErrFutureTimeout  = future.ErrTimeout
	ErrFutureCanceled = future.ErrCanceled
)

// The Dialer handles connecting to a server and creating a connection
type Dialer = transport.Dialer

// NewDialer returns a new Dialer
func NewDialer(tc *tls.Config, td time.Duration) *Dialer {
	return transport.NewDialer(transport.DialConfig{TLSConfig: tc, Timeout: td})
}

// The Launcher helps with launching a server and accepting connections
type Launcher = transport.Launcher

// NewLauncher returns a new Launcher
func NewLauncher(tc *tls.Config) *Launcher {
	return transport.NewLauncher(transport.LaunchConfig{TLSConfig: tc})
}
