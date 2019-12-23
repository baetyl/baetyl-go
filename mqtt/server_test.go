package mqtt

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestMqttTcp(t *testing.T) {
	handle := func(conn Connection, _ bool) {
		p, err := conn.Receive()
		assert.NoError(t, err)
		err = conn.Send(p, false)
		assert.NoError(t, err)
	}
	endpoints := []*Endpoint{
		&Endpoint{
			Address: "tcp://:0",
			Handle:  handle,
		},
		&Endpoint{
			Address: "tcp://127.0.0.1:0",
			Handle:  handle,
		},
	}
	cert := utils.Certificate{}
	m, err := NewTransport(endpoints, cert)
	assert.NoError(t, err)
	defer m.Close()
	time.Sleep(time.Millisecond * 100)

	// TODO: test timeout
	dailer := NewDialer(nil, time.Duration(0))
	pkt := NewConnect()
	pkt.ClientID = m.servers[0].Addr().String()
	conn, err := dailer.Dial(getURL(m.servers[0], "tcp"))
	assert.NoError(t, err)
	err = conn.Send(pkt, false)
	assert.NoError(t, err)
	res, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()

	pkt.ClientID = m.servers[1].Addr().String()
	conn, err = dailer.Dial(getURL(m.servers[1], "tcp"))
	assert.NoError(t, err)
	err = conn.Send(pkt, true)
	assert.NoError(t, err)
	res, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()
}

func TestMqttTcpTls(t *testing.T) {
	count := int32(0)
	handle := func(conn Connection, _ bool) {
		c := atomic.AddInt32(&count, 1)
		p, err := conn.Receive()
		fmt.Println(p, err)
		if c == 1 {
			assert.EqualError(t, err, "remote error: tls: bad certificate")
			assert.Nil(t, p)
			return
		}
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.True(t, IsBidirectionalAuthentication(conn))
		err = conn.Send(p, false)
		assert.NoError(t, err)
	}
	endpoints := []*Endpoint{
		&Endpoint{
			Address: "ssl://localhost:0",
			Handle:  handle,
		},
	}
	cert := utils.Certificate{
		CA:   "./testcert/ca.pem",
		Key:  "./testcert/server.key",
		Cert: "./testcert/server.pem",
	}
	m, err := NewTransport(endpoints, cert)
	assert.NoError(t, err)
	defer m.Close()
	time.Sleep(time.Millisecond * 100)

	url := getURL(m.servers[0], "ssl")
	pkt := NewConnect()
	pkt.ClientID = m.servers[0].Addr().String()

	// count: 1
	dailer := NewDialer(nil, time.Duration(0))
	conn, err := dailer.Dial(url)
	assert.Nil(t, conn)
	switch err.Error() {
	case "x509: certificate signed by unknown authority":
	case "x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs":
	default:
		assert.FailNow(t, "error expected")
	}

	// count: 2
	ctc, err := utils.NewTLSConfigClient(utils.Certificate{
		CA:                 "./testcert/ca.pem",
		Key:                "./testcert/testssl2.key",
		Cert:               "./testcert/testssl2.pem",
		InsecureSkipVerify: true,
	})
	assert.NoError(t, err)
	dailer = NewDialer(ctc, time.Duration(0))
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(pkt, false)
	assert.NoError(t, err)
	res, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()
}

func TestMqttWebSocket(t *testing.T) {
	handle := func(conn Connection, _ bool) {
		p, err := conn.Receive()
		assert.NoError(t, err)
		err = conn.Send(p, false)
		assert.NoError(t, err)
	}
	endpoints := []*Endpoint{
		&Endpoint{
			Address: "ws://localhost:0",
			Handle:  handle,
		},
		&Endpoint{
			Address: "ws://127.0.0.1:0/mqtt",
			Handle:  handle,
		},
	}
	var cert utils.Certificate
	m, err := NewTransport(endpoints, cert)
	assert.NoError(t, err)
	defer m.Close()
	time.Sleep(time.Millisecond * 100)

	dailer := NewDialer(nil, time.Duration(0))
	pkt := NewConnect()
	pkt.ClientID = m.servers[0].Addr().String()
	conn, err := dailer.Dial(getURL(m.servers[0], "ws"))
	assert.NoError(t, err)
	err = conn.Send(pkt, false)
	assert.NoError(t, err)
	res, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()

	pkt.ClientID = m.servers[1].Addr().String()
	conn, err = dailer.Dial(getURL(m.servers[1], "ws") + "/mqtt")
	assert.NoError(t, err)
	err = conn.Send(pkt, true)
	assert.NoError(t, err)
	res, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()

	pkt.ClientID = m.servers[1].Addr().String() + "-1"
	conn, err = dailer.Dial(getURL(m.servers[1], "ws") + "/notexist")
	assert.NoError(t, err)
	err = conn.Send(pkt, false)
	assert.NoError(t, err)
	res, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()
}

func TestMqttWebSocketTls(t *testing.T) {
	handle := func(conn Connection, _ bool) {
		p, err := conn.Receive()
		fmt.Println(p, err)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.True(t, IsBidirectionalAuthentication(conn))
		err = conn.Send(p, false)
		assert.NoError(t, err)
	}
	endpoints := []*Endpoint{
		&Endpoint{
			Address: "wss://localhost:0/mqtt",
			Handle:  handle,
		},
	}
	cert := utils.Certificate{
		CA:   "./testcert/ca.pem",
		Key:  "./testcert/server.key",
		Cert: "./testcert/server.pem",
	}
	m, err := NewTransport(endpoints, cert)
	assert.NoError(t, err)
	defer m.Close()
	time.Sleep(time.Millisecond * 100)

	url := getURL(m.servers[0], "wss") + "/mqtt"
	pkt := NewConnect()
	pkt.ClientID = m.servers[0].Addr().String()

	// count: 1
	dailer := NewDialer(nil, time.Duration(0))
	conn, err := dailer.Dial(url)
	assert.Nil(t, conn)
	switch err.Error() {
	case "x509: certificate signed by unknown authority":
	case "x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs":
	default:
		assert.FailNow(t, "error expected")
	}

	// count: 2
	ctc, err := utils.NewTLSConfigClient(utils.Certificate{
		CA:                 "./testcert/ca.pem",
		Key:                "./testcert/testssl2.key",
		Cert:               "./testcert/testssl2.pem",
		InsecureSkipVerify: true,
	})
	assert.NoError(t, err)
	dailer = NewDialer(ctc, time.Duration(0))
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(pkt, false)
	assert.NoError(t, err)
	res, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, pkt.String(), res.String())
	conn.Close()
}

func TestServerException(t *testing.T) {
	handle := func(conn Connection, _ bool) {
		p, err := conn.Receive()
		assert.NoError(t, err)
		err = conn.Send(p, false)
		assert.NoError(t, err)
	}
	endpoints := []*Endpoint{
		&Endpoint{
			Address: "tcp://:28767",
			Handle:  handle,
		},
		&Endpoint{
			Address: "tcp://:28767",
			Handle:  handle,
		},
	}

	var cert utils.Certificate
	_, err := NewTransport(endpoints, cert)
	switch err.Error() {
	case "listen tcp :28767: bind: address already in use":
	case "listen tcp :28767: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.":
	default:
		assert.FailNow(t, "error expected")
	}

	endpoints = []*Endpoint{
		&Endpoint{
			Address: "tcp://:28767",
			Handle:  handle,
		},
		&Endpoint{
			Address: "ssl://:28767",
			Handle:  handle,
		},
	}
	_, err = NewTransport(endpoints, cert)
	assert.EqualError(t, err, "tls: neither Certificates nor GetCertificate set in Config")

	endpoints = []*Endpoint{
		&Endpoint{
			Address: "ws://:28767/v1",
			Handle:  handle,
		},
		&Endpoint{
			Address: "wss://:28767/v2",
			Handle:  handle,
		},
	}
	_, err = NewTransport(endpoints, cert)
	assert.EqualError(t, err, "tls: neither Certificates nor GetCertificate set in Config")

	endpoints = []*Endpoint{
		&Endpoint{
			Address: "ws://:28767/v1",
			Handle:  handle,
		},
		&Endpoint{
			Address: "ws://:28767/v1",
			Handle:  handle,
		},
	}
	_, err = NewTransport(endpoints, cert)
	switch err.Error() {
	case "listen tcp :28767: bind: address already in use":
	case "listen tcp :28767: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.":
	default:
		assert.FailNow(t, "error expected")
	}

	// TODO: test more special case
	// endpoints = []string{"ws://:28767/v1", "ws://0.0.0.0:28767/v2"}
	// endpoints = []string{"ws://localhost:28767/v1", "ws://127.0.0.1:28767/v2"}
}

func getPort(s Server) string {
	_, port, _ := net.SplitHostPort(s.Addr().String())
	return port
}

func getURL(s Server, protocol string) string {
	return fmt.Sprintf("%s://%s", protocol, s.Addr().String())
}
