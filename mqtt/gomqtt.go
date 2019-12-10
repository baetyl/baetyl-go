package mqtt

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/256dpi/gomqtt/client"
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
	return packet.NewPublish()
}

// Puback the puback packet
type Puback = packet.Puback

// Subscribe the subscribe packet
type Subscribe = packet.Subscribe

// Suback the suback packet
type Suback = packet.Suback

// Unsuback the unsuback packet
type Unsuback = packet.Unsuback

// Pingreq the pingreq packet
type Pingreq = packet.Pingreq

// NewPingreq creates a new Pingreq packet.
func NewPingreq() *Pingreq {
	return packet.NewPingreq()
}

// Pingresp the pingresp packet
type Pingresp = packet.Pingresp

// Unsubscribe the unsubscribe packet
type Unsubscribe = packet.Unsubscribe

// Connect the connect packet
type Connect = packet.Connect

// NewConnect creates a new Connect packet
func NewConnect() *Connect {
	return packet.NewConnect()
}

// Connack the connack packet
type Connack = packet.Connack

// Disconnect the disconnect packet
type Disconnect = packet.Disconnect

// NewDisconnect creates a new Disconnect packet
func NewDisconnect() *Disconnect {
	return packet.NewDisconnect()
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
type Tracker = client.Tracker

// NewTracker creates a new tracker
func NewTracker(timeout time.Duration) *Tracker {
	return client.NewTracker(timeout)
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
	if tcps, ok := conn.(*transport.NetConn); ok {
		inner = tcps.UnderlyingConn()
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
	ErrClientAlreadyConnecting = client.ErrClientAlreadyConnecting
	ErrClientNotConnected      = client.ErrClientNotConnected
	ErrClientMissingID         = client.ErrClientMissingID
	ErrClientConnectionDenied  = client.ErrClientConnectionDenied
	ErrClientMissingPong       = client.ErrClientMissingPong
	ErrClientExpectedConnack   = client.ErrClientExpectedConnack
	ErrFailedSubscription      = client.ErrFailedSubscription
)
